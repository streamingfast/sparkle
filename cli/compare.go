package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/streamingfast/sparkle/diff"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/metrics"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

var compareCmd = &cobra.Command{
	Use:   "compare <table_name> <start block> <end block> <dsn1> <schema1> <dsn2> <schema2>",
	Short: "Compares entities two different databases and finds the first block at which they are different",
	Args:  cobra.ExactArgs(7),
	RunE:  runCompare,
}

func init() {
	RootCmd.AddCommand(compareCmd)
}

type diffReport struct {
	First   interface{}  `json:"first"`
	Second  interface{}  `json:"second"`
	Events  []string     `json:"differences"`
	Indexes map[int]bool `json:"-"`
}

func (dr *diffReport) String() string {
	if len(dr.Indexes) == 0 {
		output, _ := json.MarshalIndent(dr, "", "   ")
		return fmt.Sprintln(string(output))
	}

	// only output elements which have differences

	f := dr.First.([]entity.Interface)
	s := dr.Second.([]entity.Interface)

	var ft []entity.Interface
	var st []entity.Interface

	for idx := range dr.Indexes {
		ft = append(ft, f[idx])
		st = append(st, s[idx])
	}

	dr.First = ft
	dr.Second = st

	output, _ := json.MarshalIndent(dr, "", "   ")
	return fmt.Sprintln(string(output))
}

func runCompare(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	tableName := args[0]
	startBlock, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	endBlock, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	psqlDSN1 := args[3]
	psqlSchema1 := args[4]
	psqlDSN2 := args[5]
	psqlSchema2 := args[6]

	db1, err := newDatabase(ctx, psqlDSN1, psqlSchema1, tableName)
	if err != nil {
		return err
	}
	db2, err := newDatabase(ctx, psqlDSN2, psqlSchema2, tableName)
	if err != nil {
		return err
	}

	subgraphDef := subgraph.MainSubgraphDef
	entityPointerType, ok := subgraphDef.Entities.GetType(tableName)
	if !ok {
		return fmt.Errorf("could not determine type")
	}

	diffReportMap := map[int]*diffReport{}

	diffFunc := func(blockNum int) bool {
		e1, err := db1.Get(ctx, blockNum)
		if err != nil {
			panic(fmt.Errorf("%s: getting %s at blockNum %d: %w", db1, tableName, blockNum, err))
		}

		e2, err := db2.Get(ctx, blockNum)
		if err != nil {
			panic(fmt.Errorf("%s: getting %s at blockNum %d: %w", db2, tableName, blockNum, err))
		}

		var different bool
		var differences []string
		indexes := map[int]bool{}

		diff.Diff(e1, e2,
			diff.OnEvent(func(event diff.Event) {
				switch event.Kind {
				case diff.KindChanged:
					different = true

					idx, ok := event.Path.LastSliceIndex()
					if ok {
						indexes[idx] = true
					}
					differences = append(differences, event.String())
				}
			}),
			diff.CmpOption(cmp.Comparer(func(i, j string) bool {
				return strings.ToLower(i) == strings.ToLower(j)
			})),
			diff.CmpOption(cmp.Comparer(func(i, j entity.Base) bool {
				return true
			})),
			diff.CmpOption(cmp.Comparer(func(i, j entity.Int) bool {
				i1 := i.Int().Int64()
				i2 := j.Int().Int64()
				equal := i1 == i2
				return equal
			})),
			diff.CmpOption(cmp.Comparer(func(i, j *entity.Int) bool {
				if i == nil && j == nil {
					return true
				}

				if i == nil {
					jfl := j.Int().Int64()
					if math.Abs(float64(jfl)) > 0 {
						return false
					}
					return true
				}

				if j == nil {
					ifl := j.Int().Int64()
					if math.Abs(float64(ifl)) > 0 {
						return false
					}
					return true
				}

				i1 := i.Int().Int64()
				i2 := j.Int().Int64()
				equal := i1 == i2
				return equal
			})),
			diff.CmpOption(cmp.Comparer(func(i, j entity.Float) bool {
				var EPSILON float64 = 0.000000001

				f1, _ := i.Float().Float64()
				f2, _ := j.Float().Float64()
				equalFloat := math.Abs(f1-f2) <= EPSILON
				if !equalFloat {
					return false
				}

				return equalFloat
			})),
			diff.CmpOption(cmp.Comparer(func(i, j *entity.Float) bool {
				var EPSILON float64 = 0.000000001

				if i == nil && j == nil {
					return true
				}

				if i == nil {
					jfl, _ := j.Float().Float64()
					if math.Abs(jfl) > 0 {
						return false
					}
					return true
				}

				if j == nil {
					ifl, _ := i.Float().Float64()
					if math.Abs(ifl) > 0 {
						return false
					}
					return true
				}

				f1, _ := i.Float().Float64()
				f2, _ := j.Float().Float64()
				equalFloat := math.Abs(f1-f2) <= EPSILON
				if !equalFloat {
					return false
				}

				return equalFloat
			})),
			diff.CmpOption(cmpopts.IgnoreUnexported(entity.Base{}, reflect.ValueOf(entityPointerType))),
		)

		diffReportMap[blockNum] = &diffReport{
			First:   e1,
			Second:  e2,
			Events:  differences,
			Indexes: indexes,
		}
		return different
	}

	fmt.Printf("Searching for differences for %s on block range [%d, %d]...\n", tableName, startBlock, endBlock)
	block := binarySearch(startBlock, endBlock, diffFunc, 0)

	if block == -1 {
		fmt.Printf("Great news! No differences found!\n")
		return nil
	}

	fmt.Printf("Differences found @ block %d:\n", block)
	fmt.Printf("Report for block %d:\n", block)
	fmt.Printf("%s\n", diffReportMap[block])
	return nil
}

func binarySearch(minBlock, maxBlock int, isDiff func(blockNum int) bool, callDepth int) int {
	if minBlock == maxBlock {
		return maxBlock
	}

	differentAtLowerBound := isDiff(minBlock)
	if differentAtLowerBound && callDepth == 0 {
		// doomed from the start
		return minBlock
	}

	differentAtUpperBound := isDiff(maxBlock)
	if !differentAtLowerBound && !differentAtUpperBound && callDepth == 0 {
		// the same on entire range
		return -1
	}

	midpoint := int((minBlock + maxBlock) / 2)
	differentAtMidpoint := isDiff(midpoint)

	if !differentAtLowerBound && differentAtMidpoint {
		return binarySearch(minBlock, midpoint, isDiff, callDepth+1)
	} else if !differentAtMidpoint && differentAtUpperBound {
		return binarySearch(midpoint+1, maxBlock, isDiff, callDepth+1)
	} else {
		return minBlock
	}
}

type database struct {
	conn   *sqlx.DB
	schema string
	table  string

	store storage.Store
}

func (db *database) String() string {
	return fmt.Sprintf("%s@%s", db.conn.DriverName(), db.schema)
}

func (db *database) Get(ctx context.Context, blockNum int) ([]entity.Interface, error) {
	subgraphDef := subgraph.MainSubgraphDef
	model, ok := subgraphDef.Entities.GetInterface(db.table)
	if !ok {
		return nil, fmt.Errorf("table not found")
	}

	res, err := db.store.LoadAllDistinct(ctx, model, uint64(blockNum))
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].GetID() < res[j].GetID()
	})

	return res, nil
}

func newDatabase(ctx context.Context, dsnString, schema, tableName string) (*database, error) {
	dsn, err := postgres.ParseDSN(dsnString)
	if err != nil {
		return nil, err
	}

	dbConnectCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dbCancel()

	db, err := sqlx.ConnectContext(dbConnectCtx, "postgres", dsn.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	store, err := postgres.New(zap.NewNop(), metrics.NewBlockMetrics(), db, schema, "", subgraph.MainSubgraphDef, nil, true)

	return &database{
		conn:   db,
		schema: schema,
		store:  store,
		table:  tableName,
	}, nil
}
