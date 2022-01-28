package dryrunstore

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/csvexport"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/sf/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

type Store struct {
	ctx context.Context

	subgraph *subgraph.Definition
	Cache    map[string]map[string]entity.Interface

	startBlock         uint64
	endBlock           uint64
	lastBlockTimestamp time.Time
	lastBlockNum       uint64

	outputPath string

	csvExporters map[string]*csvexport.Writer

	logger *zap.Logger
}

func New(ctx context.Context, subgraphDef *subgraph.Definition, logger *zap.Logger, endBlock uint64, outputPath string) *Store {
	cache := map[string]map[string]entity.Interface{}
	for tbl := range subgraphDef.Entities.Data() {
		cache[tbl] = map[string]entity.Interface{}
	}
	return &Store{
		ctx:          ctx,
		subgraph:     subgraphDef,
		Cache:        cache,
		startBlock:   subgraphDef.StartBlock,
		endBlock:     endBlock,
		logger:       logger,
		outputPath:   outputPath,
		csvExporters: map[string]*csvexport.Writer{},
	}
}

func (s *Store) OpenOutputFiles() error {
	s.logger.Info("setting up csv exporters")
	dryrunOutput, err := dstore.NewSimpleStore(s.outputPath)
	if err != nil {
		return fmt.Errorf("new dry run store: %w", err)
	}

	for tblName := range s.subgraph.Entities.Data() {
		filename := fmt.Sprintf("%s-%010d-%010d.csv", tblName, s.startBlock, s.endBlock)

		exp, err := csvexport.New(s.ctx, dryrunOutput, filename, s.endBlock, true)
		if err != nil {
			return fmt.Errorf("new csv exporter: %w", err)
		}
		s.csvExporters[tblName] = exp
	}
	return nil
}

func (s *Store) BatchSave(ctx context.Context, block *pbcodec.Block, updates map[string]map[string]entity.Interface, cursor string) (err error) {
	for tblName, rows := range updates {
		exporter := s.csvExporters[tblName]
		cachedTable := s.Cache[tblName]

		for id, row := range rows {
			prevRow, found := cachedTable[id]
			if !found { // CREATE case
				if row == nil {
					panic("deleting something that didn't exist?")
				}

				row.SetBlockRange(&entity.BlockRange{StartBlock: block.Number})
				cachedTable[id] = row
				continue
			}

			fmt.Printf("PTR: %p %p\n", prevRow, row)

			br := prevRow.GetBlockRange()
			br.EndBlock = block.Number
			exporter.Encode(prevRow)

			if row == nil { // DELETE case
				delete(cachedTable, id)
				continue
			}

			row.SetBlockRange(&entity.BlockRange{StartBlock: block.Number})
			cachedTable[id] = row
		}
	}
	return nil
}

func (s *Store) Load(ctx context.Context, id string, out entity.Interface, blockNum uint64) error {
	// Literal CLONE of squashable.Store::Load
	tableName := entity.GetTableName(out)
	tbl, found := s.Cache[tableName]
	if !found {
		return nil
	}

	if e, found := tbl[id]; found {
		ve := reflect.ValueOf(out).Elem()
		ve.Set(reflect.ValueOf(e).Elem())
	}

	return nil
}

func (s *Store) LoadAllDistinct(ctx context.Context, model entity.Interface, blockNum uint64) ([]entity.Interface, error) {
	return nil, nil
}

func (s *Store) LoadCursor(ctx context.Context) (string, error) {
	return "", nil
}

func (s *Store) CleanDataAtBlock(ctx context.Context, blockNum uint64) error {
	return nil
}

func (s *Store) CleanUpFork(ctx context.Context, newHeadBlock uint64) error {
	return fmt.Errorf("implied irreversibility here, should never need to clean-up a fork")
}

func (s *Store) Close() error {
	// Flush all remaining entities, inspired by FlushEntities
	// on the Squashable store
	s.logger.Info("setting up flush entities store")
	for tblName := range s.subgraph.Entities.Data() {
		exporter := s.csvExporters[tblName]
		for _, row := range s.Cache[tblName] {
			if err := exporter.Encode(row); err != nil {
				return fmt.Errorf("csv encode: %w", err)
			}
		}
		if err := exporter.Close(); err != nil {
			return fmt.Errorf("csv close: %w", err)
		}
	}

	return nil
}
