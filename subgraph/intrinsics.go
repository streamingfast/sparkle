package subgraph

import (
	"time"

	"github.com/streamingfast/eth-go"
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

	// Will retry until either token is valid or RPC's HEAD block reached ours
	GetTokenInfo(address eth.Address, f TokenValidator) (token *eth.Token, valid bool)
}

type TokenValidator func(*eth.Token) bool

type BlockRef interface {
	ID() string
	Number() uint64
	Timestamp() time.Time
}
