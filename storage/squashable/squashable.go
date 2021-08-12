package squashable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

type store struct {
	ctx   context.Context
	cache map[string]map[string]entity.Interface
	step  int

	subgraph *subgraph.Definition

	entityWriters          map[string]*entityWriter
	entitiesFlushCompleted chan struct{}

	startBlock         uint64
	endBlock           uint64
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

func New(ctx context.Context, logger *zap.Logger, subgraph *subgraph.Definition, step int, startBlock, endBlock uint64) *store {
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
		startBlock:    startBlock,
		endBlock:      endBlock,
		logger:        logger,
	}
}

func (s *store) FlushEntities(store dstore.Store) {
	s.logger.Info("setting up flush entities store")
	for tblName := range s.subgraph.Entities.Data() {
		entWriter := newEntityWriter(s.logger, store, fmt.Sprintf("%s/%010d-%010d-entities.jsonl", tblName, s.startBlock, s.endBlock), tblName)
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

func (s *store) WriteSnapshot(out dstore.Store) (string, error) {

	// Purge of old entities before flushing, because we know these
	// things will not be Loaded by the next shard.
	//
	// For example:
	// * transactions that are writte and read ONLY in the same block
	// * pairHourData: which we know are read/written to only during the same hour.
	s.purgeCache()

	raw, err := json.Marshal(s.cache)
	if err != nil {
		return "", fmt.Errorf("final map marshal: %w", err)
	}
	filename := fmt.Sprintf("%010d-%010d.json", s.startBlock, s.endBlock)
	if err := out.WriteObject(context.Background(), filename, bytes.NewReader(raw)); err != nil {
		return "", fmt.Errorf("write object final map: %w", err)
	}

	return filename, nil
}

func (s *store) purgeCache() {
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

func (s *store) Preload(ctx context.Context, in dstore.Store) error {
	filesToLoad := []string{}
	var endRange uint64
	err := in.Walk(context.Background(), "", "", func(filename string) (err error) {
		startBlockNum, endBlockNum, err := getBlockRange(filename)
		if err != nil {
			return err
		}

		if endRange == 0 {
			endRange = endBlockNum
		} else {
			if startBlockNum != (endRange + 1) {
				return fmt.Errorf("broken file contiguity at %q (previous range end was %d)", filename, endRange)
			}
			endRange = endBlockNum
		}

		if startBlockNum < s.startBlock {
			filesToLoad = append(filesToLoad, filename)
		} else {
			return dstore.StopIteration
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to walk input store: %w", err)
	}

	for _, filepath := range filesToLoad {
		if err := s.loadSnapshotFile(ctx, in, filepath); err != nil {
			return fmt.Errorf("unable to load snapshot file: %w", err)
		}
	}

	return nil
}

func (s *store) loadSnapshotFile(ctx context.Context, in dstore.Store, snapshotsfilePath string) error {
	reader, err := in.OpenObject(ctx, snapshotsfilePath)
	if err != nil {
		return fmt.Errorf("unable to load input file %q: %w", snapshotsfilePath, err)
	}
	defer reader.Close()

	s.logger.Info("decoding filepath", zap.String("filepath", snapshotsfilePath))
	if err := json.NewDecoder(reader).Decode(s); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func getBlockRange(filename string) (uint64, uint64, error) {
	number := regexp.MustCompile(`(\d{10})-(\d{10})`)
	match := number.FindStringSubmatch(filename)
	if match == nil {
		return 0, 0, fmt.Errorf("no block range in filename: %s", filename)
	}

	startBlock, _ := strconv.ParseUint(match[1], 10, 64)
	stopBlock, _ := strconv.ParseUint(match[2], 10, 64)
	return startBlock, stopBlock, nil
}

func (s *store) UnmarshalJSON(in []byte) error {
	s.logger.Debug("unmarshalling file")
	var tables map[string]json.RawMessage
	if err := json.Unmarshal(in, &tables); err != nil {
		return err
	}

	for tblName, elements := range tables {
		s.logger.Debug("handling table", zap.String("table_name", tblName))
		reflectType, ok := s.subgraph.Entities.GetType(tblName)
		if !ok {
			return fmt.Errorf("no entity registered for table name %q", tblName)
		}

		var entities map[string]json.RawMessage
		if err := json.Unmarshal(elements, &entities); err != nil {
			return err
		}
		s.logger.Debug("unmarshalled entities map")
		cachedTable := s.cache[tblName]
		if cachedTable == nil {
			cachedTable = make(map[string]entity.Interface)
			s.cache[tblName] = cachedTable
		}
		s.logger.Debug("handling entities", zap.String("table_name", tblName), zap.Int("entity_count", len(entities)))
		for id, rawEntity := range entities {
			s.logger.Debug("handling entity",
				zap.String("table_name", tblName),
				zap.String("entitiy_id", id),
			)
			el := reflect.New(reflectType).Interface()
			if err := json.Unmarshal(rawEntity, el); err != nil {
				return fmt.Errorf("unmarshal raw entity: %w", err)
			}

			modifier := el.(entity.Interface) // 0xaa in the yellow box ( the entity we are reading form the current download state file)
			modifier.SetExists(true)

			cachedTable[id] = s.subgraph.MergeFunc(s.step, cachedTable[id], modifier)
		}
	}

	return nil
}

func (s *store) BatchSave(ctx context.Context, block *pbcodec.Block, updates map[string]map[string]entity.Interface, cursor string) error {
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

func (s *store) LoadAllDistinct(ctx context.Context, model entity.Interface, blockNum uint64) (out []entity.Interface, err error) {
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
