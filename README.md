# Sparkle by StreamingFast - IN DEVELOPER PREVIEW
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/sparkle)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A The Graph indexer with superpowers.

Sparkle is powered by [StreamingFast's Firehose], it runs orders of magnitude faster than the original `graph-node` implementation.

Current, the Subgraph (written in Typescript) needs to be written in Go.

# Features

* Drop-in replacement for the indexer par of Subgraphs
  * Writes directly to postgres and leaves the serving of the requests to the original `graph-node`
* Extremely fast linear processing
* Parallel processing powers to shrink processing time from months to hours.
* Richer data for your indexing needs
  * Execution traces, state changes, full EVM call trees, log events, input and return data, all available for you to inspect and index.
* Native multi-blockchain support
* Consumes the same `subgraph.yaml` you already know and love from the original The Graph implementation.


# Getting Started

Get and compile the `sparkle` binary:

```go get -v github.com/streamingfast/sparkle/cmd/sparkle```

Scaffold your project:

```sparkle codegen ./path/to/subgraph.yaml github.com/my-org/my-go-module-repo```

Update handlers that are scaffolded in `subgraph/handler_*.go`

Then run:

```bash
go run ./cmd/subgraph index --dry-run
# or build a binary + run:
go install -v ./cmd/subgraph && subgraph index --dry-run
```

Validate the output in `./dry_run`, and get ready for deployment.


# Design

## Parallel processing of subgraphs

To speed up processing of subgraphs, `sparkle` splits processes by Segments of blocks.

For example:

* Segment 1 => blocks 0 through 199,999 inclusively
* Segment 2 => blocks 200,000 through 399,999 inclusively
* ...

In a subgraph like PancakeSwap/Sushiswap/Uniswap, Segment 2 needs to know about some data that was processed in Segment 1;  for instance, it needs to know of the `Pair`s that were discovered between blocks 0 to 199,999.

To solve this, we introduce the notion of Stages, to prepare the data needed for parallel operations.

Here is an example flow and a description of what happens in each Stage's process (numbered with #)

```
           |  Segment 1  |  Segment 2   |  Segment 3
 Stage 1   |   #1        |   #2         |    #3
 Stage 2   |   #4        |   #5         |    #6
 Stage 3   |   #7        |   #8         |    #9
```

> The actual number of Segments depends on the size of the segments, and how deep the history (in blocks) subgraph you are processing is. There could be hundreds of segments for subgraph with large history. You can also adjust the number and size of segments according to the number of machines you have available, or adjust it to the target runtime (more parallelism, more speed).

> The number of stages depends on the _data dependency graph_ (see section below) of a given subgraph.

*Stage 1*:

* #1 New `Pair`s are gathered and stored
* #2 New `Pair`s discovered in Segment 2 are gathered and stored. Note: at this stage you are not aware of pairs that were discovered in Segment 1, so you cannot take action on those pairs. You need to postpone these actions to the Stage 2, where you'll have then in aggregate form.
* #3 Same as with #2

*Stage 2*:

* #4 This process starts starts with a clean slate. No previous state was accumulated
* #5 This Segment will load the data produced by all previous Segments of the previous Stage (in this case, Segment 1 of Stage 1).  For example, it can now start processing all the pairs discovered in 1), in addition to those newly (re-)discovered in this process (they were also discovered in process #2, but that's only for the use of #6.)
* #6 This Segment will load the data produced by all previous Segments of the previous Stage (in this case Segment 1 *and* 2 of Stage 1), run your aggregation methods to provide a complete view of the state needed by this Stage, using the functions you provide (see Aggregation Methods below)

*Stage 3*:

* #7 This process starts another time with a clean slate, but will execute more of its code and produce more data for the next Stage. If this is the final Stage, it will output all of the entities and their changes, exactly like the linear process.
* #8 This process will load data produced by all previous Segments of the previous Stage (in this case, Segment 1 of Stage 2), run your aggregation methods to provide a complete view of the relevant state needed by this Stage.
* #9 Similar to #8, except it will load data from Segment 1 and 2 of Stage 2 (#4 and #5).


## Staging a subgraph

In the PancakeSwap example subgraph ([which you can find here](https://github.com/streamingfast/sparkle-pancakeswap/tree/master/exchange)), most of the handler code was simply a transpilation from AssemblyScript into Go. If you look at the [Burn event handler](https://github.com/streamingfast/sparkle-pancakeswap/blob/master/exchange/handle_pair_burn_event.go) you will notice it is pretty much a clone of [the `handleBurn` method](https://github.com/pancakeswap/pancake-subgraph/blob/v2/src/exchange/core.ts#L285) in the original subgraph, with a single exception:

```golang
func (s *Subgraph) HandlePairBurnEvent(ev *PairBurnEvent) error {
	if s.StageBelow(3) {
		return nil
	}

	trx := NewTransaction(ev.Transaction.Hash.Pretty())
    ...
```

It starts with a check to know at which Stage we're at, and will simply short-circuit the code: do less processing. The runtime engine takes care of snapshotting the Entities, and loading them back in in the other Stages or Segments.


## Aggregation Methods

Each Segment of each Stage produces a dump of the *latest* state (entities and their values). If there were 1000 mutations to a UniswapFactory during Segment 1, the output of that Segment will only contain a single entry: the last version of that Entity. Only that value is useful to the next Stage's next Segment.  In the previous example, when running process #6, you will want to load data produced at #1, aggregate certain values with what was produced at #2, and start with that as the Entities available.

The `Merge` aggregation method supports several patterns to merge data between Segments:

1. Summation / averaging of numerical values (through the `Merge...()` method)

  * Ex: you use `total_transactions` to sum the number of transactions processed in Stage 3, data for each Segment will only cover what was seen during that Segment.
  * By defining something like `next.TotalTransactions.Increment(prev.TotalTransactions)` on Stage 4, you will be able to compute the summed-up value from each Segment's result.

```golang
func (next *PancakeFactory) Merge(stage int, cached *PancakeFactory) {
	if stage == 4 {
		// for summations, averaging
		next.TotalLiquidityUSD.Increment(cached.TotalLiquidityUSD)
		next.TotalLiquidityBNB.Increment(cached.TotalLiquidityBNB)
		next.TotalVolumeUSD.Increment(cached.TotalVolumeUSD)
		next.TotalVolumeBNB.Increment(cached.TotalVolumeBNB)
		next.UntrackedVolumeUSD.Increment(cached.UntrackedVolumeUSD)
	}
	return next
}
```

2. Min/max summation:

```golang
func (next *PancakeFactory) Merge(stage int, cached *PancakeFactory) {
	if stage == 4 {
        // TODO: provide example here
	}
	return next
}
```

3. Keeping track of the most recent values for certain fields. NOTE: Make sure you check that the value was properly updated on the Stage you expected it to take.

```golang
func (next *PancakeFactory) Merge(stage int, cached *PancakeFactory)  {
	// To keep only the most recent values from previous segments
	if stage == 3 && cached.MutatedOnStage == 2 {
		// Reserve0 and Reserve1 were properly set on Stage 2, so we keep them from then on.
		next.Reserve0 = cached.Reserve0
		next.Reserve1 = cached.Reserve1
	}
	return next
}
```

See the [generated source for this example](https://github.com/streamingfast/sparkle-pancakeswap/blob/master/exchange/generated.go#L1296)


## Annotation of the GraphQL schema

Here is a sample of the [PancakeSwap GraphQL](https://github.com/streamingfast/sparkle-pancakeswap/blob/master/subgraph/exchange.graphql) once annotated for parallelism:

```graphql
type PancakeFactory @entity {
  ...
  totalPairs: BigInt! @parallel(stage: 1, type: SUM)
  totalVolumeUSD: BigDecimal! @parallel(stage: 3, type: SUM)
  ...
}
type Pair @entity {
  ...
  name: String! @parallel(stage: 1)
  reserve0: BigDecimal!  @parallel(stage: 2)
  reserve1: BigDecimal!  @parallel(stage: 2)
  ...
}
```

Notice the `@parallel()` directive, with its stage number, and type. Types refer to the aggregation method defined above. Using these annotation, code is generated to automatically sum up or transform some _relative_ values (computed in a given Segment), into _absolute_ values, when they are summed up from each Segment's values, into the next stage.


### A data dependency graph

This creates a **dependency graph** of data. To be explicit:

* We first find all the Pairs in Stage 1. `totalPairs` will be able to be computed in the first run, but it will only be the count of pairs discovered _during_ that Segment. Let's call this a _relative_ value. If the first Segment discovers 5 pairs, and the second 10 pairs, on Stage 2, the generated code will sum up Segment 1 and 2's values (5 + 10) before starting Segment 3. Segment 3 will therefore be able to rely on exact and _absolute_ values for `totalPairs` (if it were to compute anything based on it). It would not be reliable to do so _within_ Stage 1, because `totalPairs` would not have a full historical view, only a partial, Segment-centric view.

* Now, before computing `totalVolumeUSD`, we will need to have `reserve0` and `reserve1` loaded properly (for those new to AMMs, dividing the reserves together give us prices). Again, we can't catch the reserves before first knowing what pairs we need to listen on. This means updating reserves on pairs need to wait to Stage 2. The dependency is: `reserve0` and `reserve1` depends on a `Pair` being created (in our example, that's [done here](https://github.com/streamingfast/sparkle-pancakeswap/blob/master/exchange/handle_factory_pair_created_event.go#L97-L112))

* At the end of all Stage+Segment (a single stage being run for a given range of blocks), a flat file (imagine a JSON with an object where tables map to a list of entities, similar to a postgres table with rows) is written with the *last* values for the Entities. The idea is to provide to the next Stage's next Segment with the latest data. In our example, it means that Segment 1 of Stage 2 would write the updated reserves, and allow Stage 3's second Segment to pick those up, and know it has legit values for `reserve0` and `reserve1`, and that there will be no missing pairs.



## Memory optimization, finalization of objects

You can imagine that holding all the state ever collected in memory when running #9 could be burdensome.

In most subgraphs however, it is known in advance that some Entities will be saved and not loaded anymore by the indexing code. That object can be declared final, so it is purged from memory and not written to the state dump for the next Stage's next Segment.

To do that, you can implement the `IsFinal` interface function, with this signature:

```golang
     func (e *MyEntity) IsFinal(blockNum uint64, blockTime time.Time) bool
```

If you return `true`, this means it is safe to assume you will not be loading that Entity anymore from within your subgraph indexing code, in any future block.  This allows to free memory, and speed up things even more.

Example:

```golang
     // In subgraph X, I know that when I saved a Transaciton, I won't need it anymore, so it's safe to mark it as Final all the time.
     func (e *Transaction) IsFinal(blockNum uint64, blockTime time.Time) bool { return true }

     // In this case, I know I will augment this Entity for the next hour's worth of blocks. Since I'll be loading it, don't finalize it.  When the time has passed, and I know I won't be reading it anymore, I can mark it as finalized, and free it from memory.
     func (p *PairHourData) IsFinal(blockNum uint64, blockTime time.Time) bool { return p.ID != fmt.Sprintf("%s-%d", p.Pair, blockTime.Unix() / 3600) }
```

# Subgraph Commands

```
$ sparkle codegen ./subgraph/exchange.yaml github.com/streamingfast/mysubgraph
$ mysubgraph create <Subgraph_NAME | mysubgraph/all>    # create  a row in `subgraph` table (current_version = nil, previous_version = nil)
$ mysubgraph deploy <Subgraph_NAME | mysubgraph/all>    #  create  a row in `subgraph_deployment` &`subgraph_version` & IPS upload & `deployment_schemas` & Update `subgraph` table current_version, previous_version (MAYBE)
$ mysubgraph inject <Subgraph_NAME | mysubgraph/all>@<VERSION>
```


## Contributing

**Issues and PR in this repo related strictly to the Sparkle application.**

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.


## License

[Apache 2.0](LICENSE)
