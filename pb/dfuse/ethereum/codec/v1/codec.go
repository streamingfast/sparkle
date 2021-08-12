package pbcodec

import (
	"encoding/hex"
	"sort"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/jsonpb"
)

func (b *Block) ID() string {
	return hex.EncodeToString(b.Hash)
}

func (b *Block) PreviousID() string {
	return hex.EncodeToString(b.Header.ParentHash)
}

func (b *Block) AsRef() bstream.BlockRef {
	return bstream.NewBlockRef(b.ID(), b.Number)
}

func (m *BigInt) MarshalJSON() ([]byte, error) {
	if m == nil {
		// FIXME: What is the right behavior regarding JSON to output when there is no bytes? Usually I think it should be omitted
		//        entirely but I'm not sure what a custom JSON marshaler can do here to convey that meaning of ok, omit this field.
		return nil, nil
	}
	return []byte(`"` + hex.EncodeToString(m.Bytes) + `"`), nil
}

func (m *BigInt) MarshalJSONPB(marshaler *jsonpb.Marshaler) ([]byte, error) {
	return m.MarshalJSON()
}

func (t *TransactionTrace) PopulateStateReverted() {
	// Calls are ordered by execution index. So the algo is quite simple.
	// We loop through the flat calls, at each call, if the parent is present
	// and reverted, the current call is reverted. Otherwise, if the current call
	// is failed, the state is reverted. In all other cases, we simply continue
	// our iteration loop.
	//
	// This works because we see the parent before its children, and since we
	// trickle down the state reverted value down the children, checking the parent
	// of a call will always tell us if the whole chain of parent/child should
	// be reverted
	//
	calls := t.Calls
	for _, call := range t.Calls {
		var parent *Call
		if call.ParentIndex > 0 {
			parent = calls[call.ParentIndex-1]
		}

		call.StateReverted = (parent != nil && parent.StateReverted) || call.StatusFailed
	}

	return
}

func (t *TransactionTrace) Logs() (logs []*Log) {
	for _, call := range t.Calls {
		if call.StateReverted {
			continue
		}

		for _, log := range call.Logs {
			logs = append(logs, log)
		}
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Index < logs[j].Index
	})

	return
}
