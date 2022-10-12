package entity

import (
	"math/big"

	"github.com/streamingfast/eth-go"
	pbcodec "github.com/streamingfast/sparkle/pb/sf/ethereum/type/v2"
)

type BaseEvent struct {
	Block       *Block
	Transaction *Transaction
}

func (e *BaseEvent) SetBlockAndTransaction(b *pbcodec.Block, t *pbcodec.TransactionTrace) {
	e.Block = newBlockFromPBBlock(b)
	e.Transaction = newTransactionFromPBTransaction(t)
}

type Block struct {
	Hash             eth.Hash
	Parent           eth.Hash
	UnclesHash       eth.Hash
	StateRoot        eth.Hash
	TransactionsRoot eth.Hash
	ReceiptsRoot     eth.Hash
	Number           uint64
	GasUsed          uint64
	GasLimit         uint64
	Timestamp        int64
	Difficulty       *big.Int
	Size             uint64
}

func newBlockFromPBBlock(b *pbcodec.Block) *Block {
	difficulty := big.Int{}
	return &Block{
		Hash:             b.Hash,
		Parent:           b.Header.ParentHash,
		UnclesHash:       b.Header.UncleHash,
		StateRoot:        b.Header.StateRoot,
		TransactionsRoot: b.Header.TransactionsRoot,
		ReceiptsRoot:     b.Header.ReceiptRoot,
		Number:           b.Number,
		GasUsed:          b.Header.GasUsed,
		GasLimit:         b.Header.GasLimit,
		Timestamp:        b.Header.Timestamp.AsTime().Unix(),
		Difficulty:       difficulty.SetBytes(b.Header.Difficulty.Bytes),
		Size:             b.GetSize(),
	}
}

type Transaction struct {
	Hash     eth.Hash
	Index    uint32
	From     eth.Address
	To       eth.Address
	Value    *big.Int
	GasUsed  uint64
	GasPrice *big.Int
	Input    []byte
}

func newTransactionFromPBTransaction(t *pbcodec.TransactionTrace) *Transaction {
	value := big.Int{}
	gasPrice := big.Int{}
	return &Transaction{
		Hash:     t.Hash,
		Index:    t.Index,
		From:     t.From,
		To:       t.To,
		Value:    value.SetBytes(t.Value.Bytes),
		GasUsed:  t.GasUsed,
		GasPrice: gasPrice.SetBytes(t.GasPrice.Bytes),
		Input:    t.Input,
	}
}

//type Call struct {
//To eth.Address
//From  eth.Address
//Block *Block
//Transaction *Transaction
//InputValues Array<EventParam>
//OutputValues Array<EventParam>
//}
///**
// * Common representation for Ethereum smart contract events.
// */
//export class Event {
//address: Address
//logIndex: BigInt
//transactionLogIndex: BigInt
//logType: string | null
//block: Block
//transaction: Transaction
//parameters: Array<EventParam>
//}
