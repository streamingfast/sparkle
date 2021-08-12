package blocks

import (
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dbin"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	pbcodec "github.com/streamingfast/sparkle/pb/dfuse/ethereum/codec/v1"
)

func init() {
	bstream.GetBlockReaderFactory = bstream.BlockReaderFactoryFunc(blockReaderFactory)
	bstream.GetBlockDecoder = bstream.BlockDecoderFunc(BlockDecoder)
	//bstream.GetProtocolFirstStreamableBlock = 0
	bstream.GetBlockWriterHeaderLen = 10
}

func BlockDecoder(blk *bstream.Block) (interface{}, error) {
	// if blk.Kind() != pbbstream.Protocol_ETH {
	//      return nil, fmt.Errorf("expected kind %s, got %s", pbbstream.Protocol_ETH, blk.Kind())
	// }

	if blk.Version() != 1 {
		return nil, fmt.Errorf("this decoder only knows about version 1, got %d", blk.Version())
	}

	block := new(pbcodec.Block)
	err := proto.Unmarshal(blk.Payload(), block)
	if err != nil {
		return nil, fmt.Errorf("unable to decode payload: %s", err)
	}

	// This whole BlockDecoder method is being called through the `bstream.Block.ToNative()`
	// method. Hence, it's a great place to add temporary data normalization calls to backport
	// some features that were not in all blocks yet (because we did not re-process all blocks
	// yet).
	//
	// Thoughts for the future: Ideally, we would leverage the version information here to take
	// a decision, like `do X if version <= 2.1` so we would not pay the performance hit
	// automatically instead of having to re-deploy a new version of bstream (which means
	// rebuild everything mostly)
	//
	// We reconstruct the state reverted value per call, for each transaction traces
	for _, trx := range block.TransactionTraces {
		trx.PopulateStateReverted()
	}

	return block, nil
}

// BlockReader reads the dbin format where each element is assumed to be a `bstream.Block`.
type BlockReader struct {
	src *dbin.Reader
}

func blockReaderFactory(reader io.Reader) (bstream.BlockReader, error) {
	return NewBlockReader(reader)
}

func NewBlockReader(reader io.Reader) (out *BlockReader, err error) {
	dbinReader := dbin.NewReader(reader)
	contentType, version, err := dbinReader.ReadHeader()
	if err != nil {
		return nil, fmt.Errorf("unable to read file header: %s", err)
	}

	Protocol := pbbstream.Protocol(pbbstream.Protocol_value[contentType])

	if Protocol != pbbstream.Protocol_ETH && version != 1 {
		return nil, fmt.Errorf("reader only knows about %s block kind at version 1, got %s at version %d", Protocol, contentType, version)
	}

	return &BlockReader{
		src: dbinReader,
	}, nil
}

func (l *BlockReader) Read() (*bstream.Block, error) {
	message, err := l.src.ReadMessage()
	if len(message) > 0 {
		pbBlock := new(pbbstream.Block)
		err = proto.Unmarshal(message, pbBlock)
		if err != nil {
			return nil, fmt.Errorf("unable to read block proto: %s", err)
		}

		blk, err := bstream.BlockFromProto(pbBlock)
		if err != nil {
			return nil, err
		}

		return blk, nil
	}

	if err == io.EOF {
		return nil, err
	}

	// In all other cases, we are in an error path
	return nil, fmt.Errorf("failed reading next dbin message: %s", err)
}
