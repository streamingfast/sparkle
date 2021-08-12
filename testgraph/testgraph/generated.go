// Code generated by sparkle. DO NOT EDIT.

package testgraph

import (
	"bytes"
	"fmt"
	"math/big"

	eth "github.com/streamingfast/eth-go"
	"github.com/streamingfast/sparkle/entity"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
	"github.com/streamingfast/sparkle/subgraph"
)

const (
	FactoryAddress = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	ZeroAddress    = "0x0000000000000000000000000000000000000000"
)

var (
	FactoryAddressBytes = eth.MustNewAddress(FactoryAddress).Bytes()
	ZeroAddressBytes    = eth.MustNewAddress(ZeroAddress).Bytes()
)

// Aliases for numerical functions
var (
	S  = entity.S
	B  = entity.B
	F  = entity.NewFloat
	FL = entity.NewFloatFromLiteral
	I  = entity.NewInt
	IL = entity.NewIntFromLiteral
	bf = func() *big.Float { return new(big.Float) }
	bi = func() *big.Int { return new(big.Int) }
)

var Definition = &subgraph.Definition{
	PackageName:         "testgraph",
	HighestParallelStep: 3,
	StartBlock:          6810753,
	IncludeFilter:       "",
	Entities: entity.NewRegistry(
		&TestEntity{},
	),
	DDL: ddl,
	Manifest: `specVersion: 0.0.2
description: Test Graph
repository: local
schema:
  file: ./testgraph.graphql
dataSources:
  - name: Factory
    network: bsc
    source:
      address: '0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73'
      abi: Factory
      startBlock: 6810753
    mapping:
      abis:
        - name: Factory
          file: ./FactoryABI.json
      eventHandlers:
        - event: PairCreated(indexed address,indexed address,address,uint256)
          handler: handlePairCreated
`,
	GraphQLSchema: `type TestEntity @entity {
  id: ID!
  name: String! @parallel(step: 2)
  set1: BigInt! @parallel(step: 1)
  set2: BigDecimal @parallel(step: 2)
  set3: String! @parallel(step: 3)
  counter1: BigInt! @parallel(step: 1, type: SUM)
  counter2: BigDecimal! @parallel(step: 2, type: SUM)
  counter3: BigInt @parallel(step: 3, type: SUM)
  derivedFromCounter1and2: BigDecimal! @parallel(step: 3)
}
`,
	Abis: map[string]string{
		"Factory": `[
  {
    "inputs": [{ "internalType": "address", "name": "_feeToSetter", "type": "address" }],
    "payable": false,
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "internalType": "address", "name": "token0", "type": "address" },
      { "indexed": true, "internalType": "address", "name": "token1", "type": "address" },
      { "indexed": false, "internalType": "address", "name": "pair", "type": "address" },
      { "indexed": false, "internalType": "uint256", "name": "", "type": "uint256" }
    ],
    "name": "PairCreated",
    "type": "event"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "INIT_CODE_PAIR_HASH",
    "outputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "name": "allPairs",
    "outputs": [{ "internalType": "address", "name": "", "type": "address" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "allPairsLength",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": false,
    "inputs": [
      { "internalType": "address", "name": "tokenA", "type": "address" },
      { "internalType": "address", "name": "tokenB", "type": "address" }
    ],
    "name": "createPair",
    "outputs": [{ "internalType": "address", "name": "pair", "type": "address" }],
    "payable": false,
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "feeTo",
    "outputs": [{ "internalType": "address", "name": "", "type": "address" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "feeToSetter",
    "outputs": [{ "internalType": "address", "name": "", "type": "address" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [
      { "internalType": "address", "name": "", "type": "address" },
      { "internalType": "address", "name": "", "type": "address" }
    ],
    "name": "getPair",
    "outputs": [{ "internalType": "address", "name": "", "type": "address" }],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
  {
    "constant": false,
    "inputs": [{ "internalType": "address", "name": "_feeTo", "type": "address" }],
    "name": "setFeeTo",
    "outputs": [],
    "payable": false,
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "constant": false,
    "inputs": [{ "internalType": "address", "name": "_feeToSetter", "type": "address" }],
    "name": "setFeeToSetter",
    "outputs": [],
    "payable": false,
    "stateMutability": "nonpayable",
    "type": "function"
  }
]
`,
	},
	New: func(base subgraph.Base) subgraph.Subgraph {
		return &Subgraph{
			Base: base,
		}
	},
	MergeFunc: func(step int, cached, new entity.Interface) entity.Interface {
		switch new.(type) {
		case interface {
			Merge(step int, new *TestEntity)
		}:
			var c *TestEntity
			if cached == nil {
				return new.(*TestEntity)
			}
			c = cached.(*TestEntity)
			el := new.(*TestEntity)
			el.Merge(step, c)
			return el
		}
		panic("unsupported merge type")
	},
}

type Subgraph struct {
	subgraph.Base
}

// TestEntity
type TestEntity struct {
	entity.Base
	Name                    string        `db:"name" csv:"name"`
	Set1                    entity.Int    `db:"set_1" csv:"set_1"`
	Set2                    *entity.Float `db:"set_2,nullable" csv:"set_2"`
	Set3                    string        `db:"set_3" csv:"set_3"`
	Counter1                entity.Int    `db:"counter_1" csv:"counter_1"`
	Counter2                entity.Float  `db:"counter_2" csv:"counter_2"`
	Counter3                *entity.Int   `db:"counter_3,nullable" csv:"counter_3"`
	DerivedFromCounter1And2 entity.Float  `db:"derived_from_counter_1_and_2" csv:"derived_from_counter_1_and_2"`
}

func NewTestEntity(id string) *TestEntity {
	return &TestEntity{
		Base:                    entity.NewBase(id),
		Set1:                    IL(0),
		Counter1:                IL(0),
		Counter2:                FL(0),
		DerivedFromCounter1And2: FL(0),
	}
}

func (_ *TestEntity) SkipDBLookup() bool {
	return false
}
func (next *TestEntity) Merge(step int, cached *TestEntity) {
	if step == 2 {
		next.Counter1 = entity.IntAdd(next.Counter1, cached.Counter1)
		if next.MutatedOnStep != 1 {
			next.Set1 = cached.Set1
		}
	}
	if step == 3 {
		next.Counter2 = entity.FloatAdd(next.Counter2, cached.Counter2)
		if next.MutatedOnStep != 2 {
			next.Name = cached.Name
			next.Set2 = cached.Set2
		}
	}
	if step == 4 {
		next.Counter3 = cached.Counter3
		if next.MutatedOnStep != 3 {
			next.Set3 = cached.Set3
			next.DerivedFromCounter1And2 = cached.DerivedFromCounter1And2
		}
	}
}

func (s *Subgraph) HandleBlock(block *pbcodec.Block) error {
	for _, trace := range block.TransactionTraces {
		logs := trace.Logs()
		for _, log := range logs {
			var ethLog interface{} = log
			eventLog := codecLogToEthLog(ethLog.(*pbcodec.Log))
			if bytes.Equal(FactoryAddressBytes, log.Address) {
				ev, err := DecodeEvent(eventLog, block, trace)
				if err != nil {
					return fmt.Errorf("parsing event: %w", err)
				}
				switch e := ev.(type) {

				case *FactoryPairCreatedEvent:
					if err := s.HandleFactoryPairCreatedEvent(e); err != nil {
						return fmt.Errorf("handling FactoryPairCreated event: %w", err)
					}
				}
			}
		}
	}
	return nil
}

func codecLogToEthLog(l *pbcodec.Log) *eth.Log {
	return &eth.Log{
		Address:    l.Address,
		Topics:     l.Topics,
		Data:       l.Data,
		Index:      l.Index,
		BlockIndex: l.BlockIndex,
	}
}

// Factory
// FactoryPairCreated event

type FactoryPairCreatedEvent struct {
	*entity.BaseEvent
	LogAddress eth.Address
	LogIndex   int

	// Fields
	Token0 eth.Address `eth:",indexed"`
	Token1 eth.Address `eth:",indexed"`
	Pair   eth.Address `eth:""`
}

var hashFactoryPairCreatedEvent = eth.Keccak256([]byte("PairCreated(address,address,address,uint256)"))

func IsFactoryPairCreatedEvent(log *eth.Log) bool {
	return bytes.Equal(log.Topics[0], hashFactoryPairCreatedEvent)
}

func NewFactoryPairCreatedEvent(log *eth.Log, block *pbcodec.Block, trace *pbcodec.TransactionTrace) (*FactoryPairCreatedEvent, error) {
	var err error
	ev := &FactoryPairCreatedEvent{
		LogAddress: log.Address,
		LogIndex:   int(log.BlockIndex),
	}

	ev.SetBlockAndTransaction(block, trace)

	dec := eth.NewLogDecoder(log)
	if _, err := dec.ReadTopic(); err != nil {
		return nil, fmt.Errorf("reading topic 0: %w", err)
	}
	f0, err := dec.ReadTypedTopic("address")
	if err != nil {
		return nil, fmt.Errorf("reading token0: %w", err)
	}
	ev.Token0 = f0.(eth.Address)
	f1, err := dec.ReadTypedTopic("address")
	if err != nil {
		return nil, fmt.Errorf("reading token1: %w", err)
	}
	ev.Token1 = f1.(eth.Address)
	ev.Pair, err = dec.DataDecoder.ReadAddress()
	if err != nil {
		return nil, fmt.Errorf("reading pair:  %w", err)
	}
	return ev, nil
}

func DecodeEvent(log *eth.Log, block *pbcodec.Block, trace *pbcodec.TransactionTrace) (interface{}, error) {

	if IsFactoryPairCreatedEvent(log) {
		ev, err := NewFactoryPairCreatedEvent(log, block, trace)
		if err != nil {
			return nil, fmt.Errorf("decoding FactoryPairCreated event: %w", err)
		}
		return ev, nil
	}

	return nil, nil
}

func (s *Subgraph) LoadDynamicDataSources(blockNum uint64) error {
	return nil
}

type DDL struct {
	createTables map[string]string
	indexes      map[string][]*index
	schemaSetup  string
}

var ddl *DDL

type index struct {
	createStatement string
	dropStatement   string
}

var createTables = map[string]string{}
var indexes = map[string][]*index{}

func init() {
	ddl = &DDL{
		createTables: map[string]string{},
		indexes:      map[string][]*index{},
	}

	Definition.DDL = ddl

	ddl.createTables["test_entity"] = `
create table if not exists %%SCHEMA%%.test_entity
(
	id text not null,

	"name" text not null,

	"set_1" numeric not null,

	"set_2" numeric,

	"set_3" text not null,

	"counter_1" numeric not null,

	"counter_2" numeric not null,

	"counter_3" numeric,

	"derived_from_counter_1_and_2" numeric not null,

	vid bigserial not null constraint test_entity_pkey primary key,
	block_range int4range not null,
	_updated_block_number numeric not null
);

alter table %%SCHEMA%%.test_entity owner to graph;
alter sequence %%SCHEMA%%.test_entity_vid_seq owned by %%SCHEMA%%.test_entity.vid;
alter table only %%SCHEMA%%.test_entity alter column vid SET DEFAULT nextval('%%SCHEMA%%.test_entity_vid_seq'::regclass);
`

	ddl.indexes["test_entity"] = func() []*index {
		var indexes []*index
		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_block_range_closed on %%SCHEMA%%.test_entity (COALESCE(upper(block_range), 2147483647)) where (COALESCE(upper(block_range), 2147483647) < 2147483647);`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_block_range_closed;`,
		})
		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_id on %%SCHEMA%%.test_entity (id);`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_id;`,
		})
		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_updated_block_number on %%SCHEMA%%.test_entity (_updated_block_number);`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_updated_block_number;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_id_block_range_fake_excl on %%SCHEMA%%.test_entity using gist (block_range, id);`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_id_block_range_fake_excl;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_name on %%SCHEMA%%.test_entity ("left"("name", 256));`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_name;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_set_1 on %%SCHEMA%%.test_entity using btree ("set_1");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_set_1;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_set_2 on %%SCHEMA%%.test_entity using btree ("set_2");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_set_2;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_set_3 on %%SCHEMA%%.test_entity ("left"("set_3", 256));`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_set_3;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_counter_1 on %%SCHEMA%%.test_entity using btree ("counter_1");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_counter_1;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_counter_2 on %%SCHEMA%%.test_entity using btree ("counter_2");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_counter_2;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_counter_3 on %%SCHEMA%%.test_entity using btree ("counter_3");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_counter_3;`,
		})

		indexes = append(indexes, &index{
			createStatement: `create index if not exists test_entity_derived_from_counter_1_and_2 on %%SCHEMA%%.test_entity using btree ("derived_from_counter_1_and_2");`,
			dropStatement:   `drop index if exists %%SCHEMA%%.test_entity_derived_from_counter_1_and_2;`,
		})

		return indexes
	}()
	ddl.schemaSetup = `
CREATE SCHEMA if not exists %%SCHEMA%%;
DO
$do$
    BEGIN
        IF NOT EXISTS (
                SELECT FROM pg_catalog.pg_roles  -- SELECT list can be empty for this
                WHERE  rolname = 'graph') THEN
            CREATE ROLE graph;
        END IF;
    END
$do$;

set statement_timeout = 0;
set idle_in_transaction_session_timeout = 0;
set client_encoding = 'UTF8';
set standard_conforming_strings = on;
select pg_catalog.set_config('search_path', '', false);
set check_function_bodies = false;
set xmloption = content;
set client_min_messages = warning;
set row_security = off;

create extension if not exists btree_gist with schema %%SCHEMA%%;


create table if not exists %%SCHEMA%%.cursor
(
	id integer not null
		constraint cursor_pkey
			primary key,
	cursor text
);
alter table %%SCHEMA%%.cursor owner to graph;

create table if not exists %%SCHEMA%%.dynamic_data_source_xxx
(
	id text not null,
	context text not null,
	abi text not null,
	vid bigserial not null
		constraint dynamic_data_source_xxx_pkey
			primary key,
	block_range int4range not null,
	_updated_block_number numeric not null
);

alter table %%SCHEMA%%.dynamic_data_source_xxx owner to graph;

create index if not exists dynamic_data_source_xxx_block_range_closed
	on %%SCHEMA%%.dynamic_data_source_xxx (COALESCE(upper(block_range), 2147483647))
	where (COALESCE(upper(block_range), 2147483647) < 2147483647);

create index if not exists dynamic_data_source_xxx_id
	on %%SCHEMA%%.dynamic_data_source_xxx (id);

create index if not exists dynamic_data_source_xxx_abi
	on %%SCHEMA%%.dynamic_data_source_xxx (abi);

`

}

func (d *DDL) InitiateSchema(handleStatement func(statement string) error) error {
	err := handleStatement(d.schemaSetup)
	if err != nil {
		return fmt.Errorf("handle statement: %w", err)
	}
	return nil
}

func (d *DDL) CreateTables(handleStatement func(table string, statement string) error) error {
	for table, statement := range d.createTables {
		err := handleStatement(table, statement)
		if err != nil {
			return fmt.Errorf("handle statement: %w", err)
		}
	}
	return nil
}

func (d *DDL) CreateIndexes(handleStatement func(table string, statement string) error) error {
	for table, idxs := range d.indexes {
		for _, idx := range idxs {
			err := handleStatement(table, idx.createStatement)
			if err != nil {
				return fmt.Errorf("handle statement: %w", err)
			}
		}
	}
	return nil
}

func (d *DDL) DropIndexes(handleStatement func(table string, statement string) error) error {
	for table, idxs := range d.indexes {
		for _, idx := range idxs {
			err := handleStatement(table, idx.dropStatement)
			if err != nil {
				return fmt.Errorf("handle statement: %w", err)
			}
		}
	}
	return nil
}
