package testgraph

import (
	"github.com/streamingfast/sparkle/entity"
	"go.uber.org/zap"
)

// Pair creation events happen at these blocks for these pairs, on BSC:
// 6810753,0x28415ff2c35b65b9e5c7de82126b4015ab9d031f
// 6810756,0x168b273278f3a8d302de5e879aa30690b7e6c28f
// 6810758,0xdd5bad8f8b360d76d12fda230f8baf42fe0022cf
// 6810760,0xb6e34b5c65eda51bb1bd4ea5f79d385fb94b9504
// 6810762,0x824eb9fadfb377394430d2744fa7c42916de3ece
// 6810764,0x7efaef62fddcca950418312c6c91aef321375a00
// 6810767,0x3dcb1787a95d2ea0eb7d00887704eebf0d79bb13
// 6810770,0x7eb5d86fd78f3852a3e0e064f2842d45a3db6ea2
// 6810772,0x74e4716e431f45807dcf19f284c7aa99f18a4fbc
// 6810775,0x61eb789d75a95caa3ff50ed7e47b96c132fec082
// 6810778,0xacf47cbeaab5c8a6ee99263cfe43995f89fb3206
// 6810780,0x16b9a82891338f9ba80e2d6970fdda79d1eb0dae
// 6810782,0x03f18135c44c64ebfdcbad8297fe5bdafdbbdd86
// 6810784,0x468b2dc8dc75990ee3e9dc0648965ad6294e7914
// 6810786,0x04eb8d58a47d2b45c9c2f673ceb6ff26e32385e3
// 6810788,0xce383277847f8217392eea98c5a8b4a7d27811b0
// 6810790,0x014608e87af97a054c9a49f81e1473076d51d9a3
// 6810792,0xd9bccbbbdfd9d67beb5d2273102ce0762421d1e3
// 6810795,0x1bdcebca3b93af70b58c41272aea2231754b23ca
// 6810797,0xd8e2f8b6db204c405543953ef6359912fe3a88d6

func (s *Subgraph) HandleFactoryPairCreatedEvent(ev *FactoryPairCreatedEvent) error {
	s.Log.Info("transaction", zap.String("trx_id", ev.Transaction.Hash.Pretty()))
	factory := NewTestEntity("1")
	if err := s.Load(factory); err != nil {
		return err
	}

	factory.Set1 = IL(int64(ev.Pair[4])) // assigns a new value, the latest should be the legit one.
	factory.Counter1 = entity.IntAdd(factory.Counter1, IL(1))

	if err := s.Save(factory); err != nil {
		return err
	}

	if s.StepBelow(2) {
		return nil
	}

	// based on the same value as `set1`
	if ev.Pair[4]%2 == 1 {
		// from time to time, we don't modify anything on Step 2,
		// so we can provoke a `MutatedOnStep` that doesn't correspond to the previous step.
		factory.Set2 = FL(float64(ev.Pair[1])).Ptr()
		factory.Counter2 = entity.FloatAdd(factory.Counter2, FL(1.0))
		if err := s.Save(factory); err != nil {
			return err
		}
	}

	if s.StepBelow(3) {
		return nil
	}

	factory.Set3 = ev.Pair.Pretty()
	if factory.Counter3 == nil {
		factory.Counter3 = IL(0).Ptr()
	}
	factory.Counter3 = entity.IntAdd(*factory.Counter3, IL(1)).Ptr()
	factory.DerivedFromCounter1And2 = entity.FloatMul(
		factory.Counter1.AsFloat(),
		factory.Counter2,
	)

	if err := s.Save(factory); err != nil {
		return err
	}

	return nil
}
