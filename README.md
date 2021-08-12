# StreamingFast Sparkle 
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/sparkle)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A TheGraph indexer with superpowers.

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

```sparkle codegen subgraph ./path/to/subgraph.yaml github.com/my-org/my-go-module-repo```

Update handlers that are scaffolded in `subgraph/handler_*.go`

Then run:

```go run ./cmd/subgraph index --dry-run```

Validate the output in `./dry_run`, and get ready for deployment.


# Advanced

## Parallel processing of subgraphs

To speed up processing of subgraphs, parallel processing allows you to chunk processing in segments of blocks

* For example: segment 1 => block 0-199999, segment 2 => block 100000-199999, etc..

In most subgraphs, for segment 2 to process properly, it needs the results of some data in segment 1.

* To solve this, we introduce the notion of steps, to prepare the data needed for parallel operations.

  * NOTE: Before running any step, all segments of previous steps need to have completed.

### Example

```
          |  Segment 1  |  Segment 2   |  Segment 3
 Step 1   |   #1        |   #2         |    #3
 Step 2   |   #4        |   #5         |    #6
 Step 3   |   #7        |   #8         |    #9
```

*Step 1*:

* #1 New trading pairs are gathered and stored
* #2 New trading pairs discovered in segment 2 are gathered and stored. Note: at this stage you are not aware of pairs that were discovered in segment 1, so you cannot take action on those pairs. You need to postpone these actions to the step 2, where you'll have then in aggregate form.
* #3 Same as with #2

*Step 2*:

* #4 This process starts starts with a clean slate. No previous state was accumulated
* #5 This segment will load the data produced by all previous segments of the previous step (in this case, Segment 1 of Step 1).  For example, it can now start processing all the pairs discovered in 1), in addition to those newly (re-)discovered in this process (they were also discovered in process #2, but that's only for the use of #6.)
* #6 This segment will load the data produced by all previous segments of the previous step (in this case Segment 1 *and* 2 of Step 1), run your aggregation methods to provide a complete view of the state needed by this step, using the functions you provide (see Aggregation Methods below)

*Step 3*:

* #7 This process starts another time with a clean slate, but will execute more of its code and produce more data for the next step. If this is the final step, it will output all of the entities and their changes, exactly like the linear process.
* #8 This process will load data produced by all previous segments of the previous step (in this case, Segment 1 of Step 2), run your aggregation methods to provide a complete view of the relevant state needed by this step.
* #9 Similar to #8, except it will load data from Segment 1 and 2 of Step 2 (#4 and #5).



## Aggregation Methods

Each segment of each step produces a dump of the *latest* state (entities and their values). If there were 1000 mutations to a UniswapFactory during segment 1, the output of that segment will only contain a single entry: the last version of that Entity. Only that value is useful to the next step's next segment.  In the previous example, when running process #6, you will want to load data produced at #1, aggregate certain values with what was produced at #2, and start with that as the Entities available.

The `Reduce` aggregation method supports several patterns to merge data between segments:

1. Summation / averaging of numerical values (through the `Reduce...()` method)

  * Ex: you use `total_transactions` to sum the number of transactions processed in step 3, data for each segment will only cover what was seen during that segment.
  * By defining something like `next.TotalTransactions.Increment(prev.TotalTransactions)` on step 4, you will be able to compute the summed-up value from each segment's result.

```golang
func (*PancakeFactory) Reduce(step int, prev, next *PancakeFactory) *PancakeFactory {
	if step == 4 {
		// for summations, averaging
		next.TotalLiquidityUSD.Increment(prev.ToatlLiquidityUSD)
		next.TotalLiquidityBNB.Increment(prev.ToatlLiquidityBNB)
		next.TotalVolumeUSD.Increment(prev.ToatlVolumeUSD)
		next.TotalVolumeBNB.Increment(prev.ToatlVolumeBNB)
		next.UntrackedVolumeUSD.Increment(prev.UntrackedVolumeUSD)
	}
	return next
}
```

2. Min/max summation:

```golang
func (*PancakeFactory) Reduce(step int, prev, next *PancakeFactory) *PancakeFactory {
	if step == 4 {
        // TODO: provide example here
	}
	return next
}
```

3. Keeping track of the most recent values for certain fields. NOTE: Make sure you check that the value was properly updated on the step you expected it to take.

```golang
func (*PancakeFactory) Reduce(step int, prev, next *PancakeFactory) *PancakeFactory  {
	// To keep only the most recent values from previous segments
	if step == 3 && prev.MutatedOnStep == 2 {
		// Reserve0 and Reserve1 were properly set on Step 2, so we keep them from then on.
		next.Reserve0 = prev.Reserve0
		next.Reserve1 = prev.Reserve1
	}
	return next
}
```


## Memory optimization, finalization of objects

You can think that holding all the state ever collected in memory when running #9 could be burdensome.

In most subgraphs however, it is known in advance that some Entities will be saved and not loaded anymore by the indexing code. That object can be declared final, so it is purged from memory and not written to the state dump for the next step's next segment.

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

->  sparkle codegen subgraph ./subgraph/exchange.yaml github.com/streamingfast/mysubgraph
->  mysubgraph create <Subgraph_NAME | mysubgraph/all> -> create  a row in `subgraph` table (current_version = nil, previsou_version = nil)
->  mysubgraph deploy <Subgraph_NAME | mysubgraph/all>-> create  a row in `subgraph_deployment` &`subgraph_version` & IPS upload & `deployment_schemas` & Update `subgraph` table current_version, previous_version (MAYBE)
->  mysubgraph inject <Subgraph_NAME | mysubgraph/all>@<VERSION>


## Contributing

**Issues and PR in this repo related strictly to the Sparkle application.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.

This codebase uses unit tests extensively, please write and run tests.

## License

[Apache 2.0](LICENSE)
