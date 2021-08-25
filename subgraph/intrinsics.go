package subgraph

import (
	"time"

	"github.com/streamingfast/sparkle/entity"
)

// Intrinsics is per subgraph and should be unique for each subgraph. The underlying implementation
// should know about its surrounding context to know when to close when at which block it's currently
// at.
//
// It's expected that the implementation will be called by one go routine at a time.
type Intrinsics interface {
	/// Entities

	Save(entity entity.Interface) error
	Load(entity entity.Interface) error
	LoadAllDistinct(model entity.Interface, blockNum uint64) ([]entity.Interface, error)
	Remove(entity entity.Interface) error

	/// Block

	// Block returns the current block being processed by your subgraph handler.
	Block() BlockRef

	/// Reproc

	StepBelow(step int) bool
	StepAbove(step int) bool

	/// JSON-RPC
	RPC(calls []*RPCCall) ([]*RPCResponse, error)
}

type BlockRef interface {
	ID() string
	Number() uint64
	Timestamp() time.Time
}

type RPCCall struct {
	ToAddr          string
	MethodSignature string // ex: "name() (string)"
}

type RPCResponse struct {
	Decoded       []interface{}
	Raw           string
	DecodingError error
	CallError     error // always deterministic
}
