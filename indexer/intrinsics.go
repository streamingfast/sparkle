package indexer

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/storage"
	"github.com/streamingfast/sparkle/subgraph"
	"go.uber.org/zap"
)

type privateIntrinsic interface {
	subgraph.Intrinsics

	startBlock(block *pbcodec.Block, step int)
	flushBlock(cursor string) error
	setStep(step int)

	loadCursor() (string, error)
	cleanStoreAtBlock(startBlock uint64) error
	cleanUpFork(longestChainStartBlock uint64) error
}

var _ privateIntrinsic = (*defaultIntrinsic)(nil)

type defaultIntrinsic struct {
	ctx       context.Context
	rpcClient *rpc.Client
	store     storage.Store

	enablePOI    bool
	aggregatePOI bool
	networkName  string

	step int

	block    *pbcodec.Block
	blockRef *blockRef

	// cached entities
	current map[string]map[string]entity.Interface
	updates map[string]map[string]entity.Interface
}

//newDefaultIntrinsic does not set the store right away, set it afterwards
func newDefaultIntrinsic(ctx context.Context, step int, rpcClient *rpc.Client) *defaultIntrinsic {
	return &defaultIntrinsic{
		ctx:       ctx,
		rpcClient: rpcClient,
		step:      step,
		enablePOI: false,

		current: make(map[string]map[string]entity.Interface),
		updates: make(map[string]map[string]entity.Interface),
	}
}

func (d *defaultIntrinsic) Save(ent entity.Interface) error {
	tableName := entity.GetTableName(ent)
	tbl, found := d.updates[tableName]
	if !found {
		tbl = make(map[string]entity.Interface)
		d.updates[tableName] = tbl
	}

	// WARN: what's the impact of setting this is we HAVEN'T
	// fetched it from the DB. Don't we rely on the `Exists()` to
	// mean "it exists in the DB" ?
	ent.SetExists(true)
	ent.SetMutated(d.step)

	tbl[ent.GetID()] = ent

	return nil
}

func (d *defaultIntrinsic) Load(ent entity.Interface) error {
	tableName := entity.GetTableName(ent)
	id := ent.GetID()
	zlog.Debug("loading entity",
		zap.String("id", id),
		zap.String("table", tableName),
		zap.Uint64("vid", ent.GetVID()),
	)
	if id == "" {
		return fmt.Errorf("id was not set before calling load")
	}

	// First check from updates
	updateTbl, found := d.updates[tableName]
	if !found {
		updateTbl = make(map[string]entity.Interface)
		d.updates[tableName] = updateTbl
	}

	cachedEntity, found := updateTbl[id]
	if found {
		if cachedEntity == nil {
			return nil
		}
		ve := reflect.ValueOf(ent).Elem()
		ve.Set(reflect.ValueOf(cachedEntity).Elem())
		return nil
	}

	// Load from DB otherwise
	currentTbl, found := d.current[tableName]
	if !found {
		currentTbl = make(map[string]entity.Interface)
		d.current[tableName] = currentTbl
	}

	cachedEntity, found = currentTbl[id]
	if found {
		if cachedEntity == nil {
			return nil
		}
		ve := reflect.ValueOf(ent).Elem()
		ve.Set(reflect.ValueOf(cachedEntity).Elem())
		return nil
	}

	if err := d.store.Load(d.ctx, id, ent, d.blockRef.Number()); err != nil {
		return fmt.Errorf("failed loading entity: %w", err)
	}

	if ent.Exists() {
		reflectType, ok := subgraph.MainSubgraphDef.Entities.GetType(tableName)
		if !ok {
			return fmt.Errorf("unable to retrieve entity type")
		}
		clone := reflect.New(reflectType).Interface()
		ve := reflect.ValueOf(clone).Elem()
		ve.Set(reflect.ValueOf(ent).Elem())
		currentTbl[id] = clone.(entity.Interface)
	} else {
		currentTbl[id] = nil
	}

	return nil
}

func (d *defaultIntrinsic) LoadAllDistinct(model entity.Interface, blockNum uint64) ([]entity.Interface, error) {
	return d.store.LoadAllDistinct(d.ctx, model, blockNum)
}

func (d *defaultIntrinsic) Remove(e entity.Interface) error {
	tableName := entity.GetTableName(e)

	tbl, found := d.updates[tableName]
	if !found {
		tbl = make(map[string]entity.Interface)
		d.updates[tableName] = tbl
	}

	tbl[e.GetID()] = nil
	return nil
}

func (d *defaultIntrinsic) Block() subgraph.BlockRef {
	return d.blockRef
}

func (d *defaultIntrinsic) StepBelow(step int) bool {
	return d.step < step
}

func (d *defaultIntrinsic) StepAbove(step int) bool {
	return d.step > step
}

func (d *defaultIntrinsic) GetTokenInfo(address eth.Address, validate subgraph.TokenValidator) (out *eth.Token, valid bool) {
	if validate == nil {
		panic("you must give a validator to intrinsic GetTokenInfo")
	}

	var sleep time.Duration
	for {
		time.Sleep(sleep)
		sleep = 500 * time.Millisecond

		var headBlockNum uint64
		var err error
		out, headBlockNum, err = d.rpcClient.GetTokenInfo(address)
		if err != nil {
			zlog.Warn("retrying GetTokenInfo on RPC error", zap.Error(err), zap.Stringer("address", address))
			continue
		}

		// with validator, we can exit early
		if validate(out) {
			return out, true
		}

		// we wait until RPC head block has reached ours+1, to be on the safe side
		if headBlockNum < d.blockRef.num {
			zlog.Info("retrying GetTokenInfo, waiting for RPC peer to be in sync", zap.Error(err), zap.Uint64("checked_head_block_num", headBlockNum), zap.Uint64("expected_head_block_num", d.blockRef.num))
			continue
		}
		return out, false
	}
}

func (d *defaultIntrinsic) setStep(step int) {
	d.step = step
}

func (d *defaultIntrinsic) startBlock(block *pbcodec.Block, step int) {
	d.block = block
	d.blockRef = asBlockRef(block)

	d.step = step

	d.current = make(map[string]map[string]entity.Interface)
	d.updates = make(map[string]map[string]entity.Interface)
}

func (d *defaultIntrinsic) flushBlock(cursor string) error {
	if d.enablePOI {
		zlog.Debug("generating poi", zap.Stringer("block", d.block))
		poi, err := d.generatePOI()
		if err != nil {
			return fmt.Errorf("unable to generate POI")
		}

		err = d.Save(poi)
		if err != nil {
			return fmt.Errorf("unable to save generated POI: %w", err)
		}
	}

	return d.store.BatchSave(d.ctx, d.block, d.updates, cursor)
}

func (d *defaultIntrinsic) generatePOI() (*entity.POI, error) {

	poi := entity.NewPOI(d.networkName)
	if err := d.Load(poi); err != nil {
		return nil, err
	}

	if !d.aggregatePOI {
		poi.Clear() // discard md5 and digest information...
	}

	if err := computePOI(poi, d.updates, d.Block()); err != nil {
		return nil, err
	}

	return poi, nil
}

func (d *defaultIntrinsic) loadCursor() (string, error) {
	return d.store.LoadCursor(d.ctx)
}

func (d *defaultIntrinsic) cleanStoreAtBlock(startBlock uint64) error {
	return d.store.CleanDataAtBlock(d.ctx, startBlock)
}

func (d *defaultIntrinsic) cleanUpFork(longestChainStartBlock uint64) error {
	return d.store.CleanUpFork(d.ctx, longestChainStartBlock)
}

type blockRef struct {
	id        string
	num       uint64
	timestamp time.Time
}

func (b blockRef) ID() string {
	return b.id
}

func (b blockRef) Number() uint64 {
	return b.num
}

func (b blockRef) Timestamp() time.Time {
	return b.timestamp
}

func asBlockRef(block *pbcodec.Block) *blockRef {
	return &blockRef{id: block.ID(), num: block.Number, timestamp: block.Header.Timestamp.AsTime()}
}

func computePOI(poi *entity.POI, updates map[string]map[string]entity.Interface, blockRef subgraph.BlockRef) error {
	count := 0

	// FIXME poi.digest must be nil always on steps 1-?4?
	previousPOIDigest := poi.Digest
	poi.Clear()

	tblNames := make([]string, 0, len(updates))
	for k := range updates {
		tblNames = append(tblNames, k)
	}
	sort.Strings(tblNames)

	for _, tblName := range tblNames {
		tblUpdates := updates[tblName]
		rowIDs := make([]string, 0, len(tblUpdates))
		for k := range tblUpdates {
			rowIDs = append(rowIDs, k)
		}
		sort.Strings(rowIDs)

		for _, id := range rowIDs {
			row := tblUpdates[id]
			count++
			err := poi.Write(tblName, id, row)
			if err != nil {
				return fmt.Errorf("unable to write entity in POI: %w", err)
			}
		}
	}
	zlog.Debug("encoded update in point", zap.Int("update_count", count))

	err := poi.Write("blocks", poi.ID, blockRef)
	if err != nil {
		return fmt.Errorf("unable to write block ref POI: %w", err)
	}

	poi.Apply()
	if previousPOIDigest != nil { // we are aggregating
		poi.AggregateDigest(previousPOIDigest)
	}
	return nil
}
