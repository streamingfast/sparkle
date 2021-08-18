package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/streamingfast/sparkle/subgraph"

	"github.com/abourget/llerrgroup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/csvexport"
	"github.com/streamingfast/sparkle/entity"
	"go.uber.org/zap"
)

var parallelToCSVCmd = &cobra.Command{
	Use:   "to-csv",
	Short: "Convert entities into CSV files for massive injection",
	Args:  cobra.NoArgs,
	RunE:  runParallelToCSV,
}

func init() {
	parallelToCSVCmd.Flags().Int("chunk-size", 0, "chunk of blocks in each outputted CSV file")
	parallelToCSVCmd.Flags().StringSlice("only-tables", []string{}, "Converts entities only for the following tables")

	parallelCmd.AddCommand(parallelToCSVCmd)
}

func runParallelToCSV(cmd *cobra.Command, _ []string) error {

	inputPath := viper.GetString("parallel-cmd-input-path")
	outputPath := viper.GetString("parallel-cmd-output-path")
	chunkSize := viper.GetInt("parallel-to-csv-cmd-chunk-size")
	startBlockNum := viper.GetUint64("parallel-cmd-start-block")
	stopBlockNum := viper.GetUint64("parallel-cmd-stop-block")
	onlyTables := viper.GetStringSlice("parallel-to-csv-cmd-only-tables")

	zlog.Info("turning entities into csv files",
		zap.String("input_path", inputPath),
		zap.String("output_path", outputPath),
		zap.Int("chunk_size", chunkSize),
		zap.Strings("only_tables", onlyTables),
		zap.Uint64("stop_block_num", stopBlockNum),
		zap.Uint64("start_block_num", startBlockNum),
		zap.Strings("only_tables", onlyTables),
	)

	onlyTablesMap := map[string]bool{}
	for _, tbl := range onlyTables {
		onlyTablesMap[tbl] = true
	}

	ctx := cmd.Context()

	if chunkSize == 0 {
		return fmt.Errorf("--chunk-size cannot be 0")
	}

	inputStore, err := dstore.NewStore(inputPath, "", "", false)
	if err != nil {
		return fmt.Errorf("unable to create input store: %w", err)
	}

	outputStore, err := dstore.NewStore(outputPath, "", "", false)
	if err != nil {
		return fmt.Errorf("unable to create output store: %w", err)
	}

	entitiesRegistry := subgraph.MainSubgraphDef.Entities
	eg := llerrgroup.New(20)
	shardStats := map[string][]zap.Field{}
	t0 := time.Now()
	for tableName := range entitiesRegistry.Data() {
		zlog.Info("processing table", zap.String("table_name", tableName))
		if eg.Stop() {
			continue // short-circuit the loop if we got an error
		}

		if len(onlyTablesMap) != 0 && !onlyTablesMap[tableName] {
			zlog.Info("skipping table", zap.String("table_name", tableName))
			continue
		}

		// SECOND ITERATION OF THIS THING: flush a snapshot of the
		// table's current view (what we have in memory, squashed and
		// purged from all the previous entities, not yet written to
		// the csv file), so we can pick up the csv writing from
		// different heights.

		tableName := tableName

		zlog.Info("starting shard", zap.String("tableName", tableName))

		eg.Go(func() error {
			ts := newTableShard(ctx, tableName, entitiesRegistry, inputStore, outputStore, chunkSize, startBlockNum, stopBlockNum)
			metrics, err := ts.Run()
			if err != nil {
				return err
			}

			shardStats[tableName] = []zap.Field{
				zap.String("table_name", tableName),
				zap.Uint64("processed_entities", metrics.entityCount),
				zap.Uint64("processed_blocks", metrics.blockCount),
				zap.Duration("duration", time.Since(metrics.startedAt)),
			}
			return nil
		})
	}

	zlog.Info("waiting for all shards")
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("processEntity shards: %w", err)
	}

	zlog.Info("all shards done", zap.Duration("total", time.Since(t0)))
	for tbl, fields := range shardStats {
		zlog.Info(fmt.Sprintf("shard completed - %s", tbl), fields...)
	}
	return nil
}

type TableShard struct {
	ctx            context.Context
	tableName      string
	in             dstore.Store
	out            dstore.Store
	chunkSize      uint64
	stopBlockNum   uint64
	csvExporter    *csvexport.Writer
	entities       map[string]entity.Interface
	metrics        *tableShardMetrics
	startBlockNum  uint64
	entityRegistry *entity.Registry
}

type tableShardMetrics struct {
	entityCount      uint64
	blockCount       uint64
	startedAt        time.Time
	fileCount        uint64
	lastProcessBlock *entity.ExportedEntities
}

func (ts *tableShardMetrics) shouldPurge() bool {
	return ts.entityCount%4000 == 0
}

func (ts *tableShardMetrics) showProgress() bool {
	return ts.entityCount%4000 == 0
}

func newTableShard(ctx context.Context, tableName string, entityRegistry *entity.Registry, in, out dstore.Store, chunkSize int, startBlockNum, stopBlockNum uint64) *TableShard {
	return &TableShard{
		ctx:            ctx,
		tableName:      tableName,
		in:             in,
		out:            out,
		chunkSize:      uint64(chunkSize),
		startBlockNum:  startBlockNum,
		stopBlockNum:   stopBlockNum,
		entityRegistry: entityRegistry,
		entities:       make(map[string]entity.Interface),
		metrics: &tableShardMetrics{
			entityCount: 0,
			blockCount:  0,
		},
	}
}

func (ts *TableShard) hasActiveCVSExporter(blockNum uint64) bool {
	if ts.csvExporter == nil {
		return false
	}
	if blockNum > ts.csvExporter.StopBlock {
		return false
	}
	return true
}

func (ts *TableShard) setupNewCSVExporter(blockNum uint64) error {
	if ts.csvExporter != nil {
		if err := ts.csvExporter.Close(); err != nil {
			return fmt.Errorf("unblae to close active csv exporter: %w", err)
		}
	}

	startBlockNum := blockNum - (blockNum % ts.chunkSize)
	endBlockNum := startBlockNum + ts.chunkSize
	filename := fmt.Sprintf("%s/%010d-%010d.csv", ts.tableName, startBlockNum, endBlockNum)

	var err error
	ts.csvExporter, err = csvexport.New(ts.ctx, ts.out, filename, endBlockNum, false)
	if err != nil {
		return fmt.Errorf("unable to create new csv exporter: %w", err)
	}

	return nil
}

func (ts *TableShard) writeEntity(blockNum uint64, ent entity.Interface) error {
	if !ts.hasActiveCVSExporter(blockNum) {
		zlog.Debug("no active csv exporter creating a new one", zap.String("table", ts.tableName), zap.Uint64("block_num", blockNum))
		if err := ts.setupNewCSVExporter(blockNum); err != nil {
			return fmt.Errorf("unable to create new csve exporter: %w", err)
		}
	}

	if err := ts.csvExporter.Encode(ent); err != nil {
		return fmt.Errorf("failed to encode csv to file: %w", err)
	}
	return nil
}

func (ts *TableShard) Run() (*tableShardMetrics, error) {
	ts.metrics.startedAt = time.Now()
	entitiesToLoad := []string{}
	var endRange uint64
	zlog.Info("retrieving relevant entity files", zap.String("table_name", ts.tableName))
	fileCount := 0
	err := ts.in.Walk(context.Background(), ts.tableName+"/", "", func(filename string) (err error) {
		fileCount++
		startBlockNum, endBlockNum, err := getBlockRange(filename)
		if err != nil {
			return fmt.Errorf("fail reading block range in %q: %w", filename, err)
		}

		if ts.stopBlockNum != 0 && startBlockNum >= ts.stopBlockNum {
			return dstore.StopIteration
		}

		if endRange == 0 {
			endRange = endBlockNum
		} else {
			if startBlockNum != (endRange + 1) {
				return fmt.Errorf("broken file contiguity at %q (previous range end was %d)", filename, endRange)
			}
			endRange = endBlockNum
		}

		if endBlockNum <= ts.startBlockNum {
			return nil
		}

		if strings.Contains(filename, "entities") {
			entitiesToLoad = append(entitiesToLoad, filename)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to walk entity files %q: %w", ts.tableName, err)
	}
	zlog.Info("found entities file to export",
		zap.String("table_name", ts.tableName),
		zap.Int("entity_file_seen_count", fileCount),
		zap.Int("entity_file_to_load", len(entitiesToLoad)),
	)

	for idx, filename := range entitiesToLoad {
		if err := ts.processEntityFile(filename); err != nil {
			return nil, fmt.Errorf("error processing file: %w", err)
		}

		if idx%10 == 0 {
			zlog.Info("entity file completed (1/10)",
				zap.String("filename", filename),
				zap.Uint64("block_count", ts.metrics.blockCount),
				zap.Uint64("entity_count", ts.metrics.entityCount),
				zap.Int("file_count", idx),
			)
		}

	}

	// FUTURE:
	// TODO: flush the snapshot to DISK so we pick back uph ere, or
	// FLUSH the contents to CSV so we're ready to pursue the
	// real-time machinery.

	if ts.metrics.lastProcessBlock != nil {
		for _, ent := range ts.entities {
			if ent == nil {
				continue
			}

			if err := ts.writeEntity(ts.metrics.lastProcessBlock.BlockNum, ent); err != nil {
				return nil, fmt.Errorf("write csv encoded: %w", err)
			}
		}

		if err := ts.csvExporter.Close(); err != nil {
			return nil, fmt.Errorf("final csv close: %w", err)
		}
	}
	return ts.metrics, nil
}

func (ts *TableShard) processEntityFile(filename string) error {
	ts.metrics.fileCount++
	zlog.Debug("processing entity file", zap.String("filename", filename), zap.String("table_name", ts.tableName))
	reader, err := ts.in.OpenObject(ts.ctx, filename)
	if err != nil {
		return fmt.Errorf("unable to load entitis file %q: %w", filename, err)
	}
	bufReader := bufio.NewReader(reader)

	for {
		ts.metrics.blockCount++
		line, err := bufReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("unable to read newline: %w", err)
		}

		currentBlock := &entity.ExportedEntities{EntityName: ts.tableName, TypeGetter: ts.entityRegistry}
		if err := json.Unmarshal([]byte(line), &currentBlock); err != nil {
			return fmt.Errorf("unable to unmarshal exported entities %q: %w", filename, err)
		}
		ts.metrics.lastProcessBlock = currentBlock

		if ts.stopBlockNum != 0 && currentBlock.BlockNum > ts.stopBlockNum {
			fmt.Printf("Interrupting at %s, %d blocks [%d total entities] [%d file count] [untilBlock %d]\n", filename, ts.metrics.blockCount, ts.metrics.blockCount, ts.metrics.fileCount, ts.stopBlockNum)
			return nil
		}

		for id, ent := range currentBlock.Entities {
			ts.metrics.entityCount++

			if sanitizableEntity, ok := ent.(entity.Sanitizable); ok {
				sanitizableEntity.Sanitize()
			}

			prev, found := ts.entities[id]

			if found {
				if prev.GetBlockRange() == nil {
					zlog.Warn("block range is nil for a seen entity",
						zap.Uint64("block_num", currentBlock.BlockNum),
						zap.String("table_name", ts.tableName),
						zap.String("entity_id", id),
					)
					continue
				}
				prev.GetBlockRange().EndBlock = currentBlock.BlockNum
				prev.SetUpdatedBlockNum(currentBlock.BlockNum)

				if ts.tableName == "poi2$" {
					ent = processProofOfIndex(prev, ent)
				}

				if err := ts.writeEntity(currentBlock.BlockNum, prev); err != nil {
					return fmt.Errorf("write csv encoded: %w", err)
				}
			}

			if ent != nil {
				ent.SetUpdatedBlockNum(currentBlock.BlockNum)
				ts.entities[id] = ent
			}

			if ts.metrics.shouldPurge() {
				for id, ent := range ts.entities {
					if purgeableEntity, ok := ent.(entity.Finalizable); ok {
						if purgeableEntity.IsFinal(currentBlock.BlockNum, currentBlock.BlockTimestamp) {
							if ent != nil {
								if err := ts.writeEntity(currentBlock.BlockNum, ent); err != nil {
									return fmt.Errorf("write csv encoded: %w", err)
								}
							}
							delete(ts.entities, id)
						}
					} else {
						break
					}
				}
			}
			if ts.metrics.showProgress() {
				zlog.Info("entities progress",
					zap.Uint64("last_block_num", currentBlock.BlockNum),
					zap.String("table_name", ts.tableName),
					zap.Uint64("entity_count", ts.metrics.entityCount),
					zap.Duration("elasped_time", time.Since(ts.metrics.startedAt)),
					zap.Int("entities_map_size", len(ts.entities)),
					zap.String("table_name", ts.tableName),
				)
			}
		}
	}
	return nil
}

func processProofOfIndex(prev, cur entity.Interface) entity.Interface {
	previous, ok := prev.(*entity.POI)
	if !ok {
		return cur
	}

	current, ok := cur.(*entity.POI)
	if !ok {
		return cur
	}
	current.AggregateDigest(previous.Digest)
	return current
}
