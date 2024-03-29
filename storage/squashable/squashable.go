package squashable

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/sf/ethereum/type/v2"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

//	const ScannerMaxCapacity   = 655360

type store struct {
	ctx   context.Context
	cache map[string]map[string]entity.Interface
	step  int

	subgraph *subgraph.Definition

	entityWriters          map[string]*entityWriter
	entitiesFlushCompleted chan struct{}

	StartBlock         uint64
	EndBlock           uint64
	lastBlockTimestamp time.Time
	lastBlockNum       uint64

	logger *zap.Logger
}

type entityWriter struct {
	filename string
	tblName  string
	store    dstore.Store
	reader   *io.PipeReader
	writer   *io.PipeWriter
	done     chan struct{}
	logger   *zap.Logger
}

func newEntityWriter(logger *zap.Logger, store dstore.Store, filename, tblName string) *entityWriter {
	reader, writer := io.Pipe()
	return &entityWriter{
		filename: filename,
		tblName:  tblName,
		store:    store,
		reader:   reader,
		writer:   writer,
		done:     make(chan struct{}),
		logger:   logger,
	}
}

func (ew *entityWriter) run(ctx context.Context) error {
	ew.logger.Info("starting to write entities from reader",
		zap.String("filename", ew.filename),
		zap.String("entity_type", ew.tblName),
	)
	err := ew.store.WriteObject(ctx, ew.filename, ew.reader)
	if err != nil {
		return fmt.Errorf("failed to write in entity table %q: %w", ew.filename, err)
	}
	close(ew.done)
	return nil
}

func (ew *entityWriter) write(data []byte) (int, error) {
	return ew.writer.Write(data)
}

func (ew *entityWriter) close() error {
	if err := ew.writer.Close(); err != nil {
		return fmt.Errorf("unable to close entities %q pipe writter: %w", ew.tblName, err)
	}
	return nil
}

func New(
	ctx context.Context,
	logger *zap.Logger,
	subgraph *subgraph.Definition,
	step int,
	startBlock, endBlock uint64,
) *store {
	cache := map[string]map[string]entity.Interface{}
	for tbl := range subgraph.Entities.Data() {
		cache[tbl] = map[string]entity.Interface{}
	}
	return &store{
		ctx:           ctx,
		subgraph:      subgraph,
		cache:         cache,
		entityWriters: map[string]*entityWriter{},
		step:          step,
		StartBlock:    startBlock,
		EndBlock:      endBlock,
		logger:        logger,
	}
}

func (s *store) FlushEntities(store dstore.Store) {
	s.logger.Info("setting up flush entities store")
	for tblName := range s.subgraph.Entities.Data() {
		entWriter := newEntityWriter(
			s.logger,
			store,
			fmt.Sprintf("%s/%010d-%010d-entities.jsonl", tblName, s.StartBlock, s.EndBlock),
			tblName,
		)
		go func() {
			err := entWriter.run(s.ctx)
			if err != nil {
				// TODO: should be a clean shutdown
				panic(fmt.Errorf("entities writer failed: %w", err))
			}
		}()
		s.entityWriters[tblName] = entWriter
	}
}

func (s *store) CloseEntities() error {
	for _, ew := range s.entityWriters {
		if err := ew.close(); err != nil {
			return err
		}
	}

	for _, ew := range s.entityWriters {
		s.logger.Info("waiting on entities flush completion signal",
			zap.String("entity_table", ew.tblName),
			zap.String("filename", ew.filename),
		)
		<-ew.done
	}
	return nil
}

func (s *store) GetStep() int {
	return s.step
}

func (s *store) GetCache() map[string]map[string]entity.Interface {
	return s.cache
}

type snapshotEntity struct {
	TableIdx int              `json:"t"`
	Entity   entity.Interface `json:"d"`
}

func (s *store) WriteSnapshot(out dstore.Store, filename string) (string, error) {
	// Purge of old entities before flushing, because we know these
	// things will not be Loaded by the next shard.
	//
	// For example:
	// * transactions that are writter and read ONLY in the same block
	// * pairHourData: which we know are read/written to only during the same hour.
	s.purgeCache()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	pr, pw := io.Pipe()
	go func() {
		defer cancel()
		if err := out.WriteObject(ctx, filename, pr); err != nil {
			s.logger.Error("snapshot: writing to storage", zap.Error(err))
			if err := pr.CloseWithError(err); err != nil {
				s.logger.Error("snapshot: closing pipe reader", zap.Error(err))
			}
		}
		close(done)
	}()
	enc := json.NewEncoder(pw)
	defer func() {
		pw.Close()
		s.logger.Info("waiting for the pipe writer to close...")
		<-done
	}()

	tableIdx := make(map[int]string)
	tableRevIdx := make(map[string]int)
	i := 0
	for tableName := range s.cache {
		tableIdx[i] = tableName
		tableRevIdx[tableName] = i
		i++
	}
	if err := enc.Encode(tableIdx); err != nil {
		return "", err
	}

	for table, entities := range s.cache {
		tidx := tableRevIdx[table]
		for _, entity := range entities {
			se := snapshotEntity{
				TableIdx: tidx,
				Entity:   entity,
			}
			if err := enc.Encode(se); err != nil {
				return "", err
			}
		}
	}

	return filename, nil
}

func (s *store) purgeCache() {
	if s.lastBlockNum == 0 {
		s.logger.Info("skipping purge_cache because no block has been processed yet")
		return
	}
	// we're at bf()OCK
	for _, rows := range s.cache {
		for id, ent := range rows {
			if purgeableEntity, ok := ent.(entity.Finalizable); ok {
				if purgeableEntity.IsFinal(s.lastBlockNum, s.lastBlockTimestamp) {
					delete(rows, id)
				}
			}
		}
	}
}

func (s *store) BatchSave(
	ctx context.Context,
	block *pbcodec.Block,
	updates map[string]map[string]entity.Interface,
	cursor string,
) error {
	s.lastBlockTimestamp = block.Header.Timestamp.AsTime()
	s.lastBlockNum = block.Number
	// naviate updates
	// if in cache update cache stop block
	for tblName, rows := range updates {
		for id, row := range rows {
			if row == nil {
				delete(s.cache[tblName], id)
			} else {
				row.SetBlockRange(&entity.BlockRange{StartBlock: block.Number})
				s.cache[tblName][id] = row
			}
		}

		if entWriter, found := s.entityWriters[tblName]; found {
			raw, err := json.Marshal(&entity.ExportedEntities{
				BlockNum:       block.Number,
				BlockTimestamp: block.Header.Timestamp.AsTime(),
				EntityName:     tblName,
				Entities:       rows,
			})
			if err != nil {
				return fmt.Errorf("unable to marshal exported entities for table %q: %w", tblName, err)
			}
			s.logger.Info("writing full block entity to pipe",
				zap.Uint64("block_num", block.Number),
				zap.String("table_name", tblName),
			)

			if _, err := entWriter.write(append(raw, '\n')); err != nil {
				return fmt.Errorf("writting new line: %w", err)
			}
		}
	}

	return nil
}

func (s *store) CleanDataAtBlock(ctx context.Context, blockNum uint64) error {
	return nil
}

func (s *store) Load(ctx context.Context, id string, out entity.Interface, blockNum uint64) error {
	tableName := entity.GetTableName(out)
	tbl, found := s.cache[tableName]
	if !found {
		return nil
	}

	if e, found := tbl[id]; found {
		ve := reflect.ValueOf(out).Elem()
		ve.Set(reflect.ValueOf(e).Elem())
	}

	return nil
}

func (s *store) LoadAllDistinct(
	ctx context.Context,
	model entity.Interface,
	blockNum uint64,
) (out []entity.Interface, err error) {
	tableName := entity.GetTableName(model)
	tbl, found := s.cache[tableName]
	if !found {
		return
	}

	for _, e := range tbl {
		out = append(out, e)
	}
	return
}

func (s *store) LoadCursor(ctx context.Context) (string, error) { return "", nil }

func (s *store) SetupTableForForkHandling(ctx context.Context) error        { return nil }
func (s *store) CleanUpFork(ctx context.Context, newHeadBlock uint64) error { return nil }
func (s *store) Close() error                                               { return nil }
