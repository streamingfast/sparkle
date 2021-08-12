#!/bin/bash

export INFO=.*

rm -rf ./dry_run/
go run ./cmd/testgraph index none@none --dry-run --dry-run-blocks 10 --dry-run-output ./dry_run

csvlook -I dry_run/*.csv

if [ ! -f "./testblocks/0006810700.dbin.zst" ]; then
    mkdir -p testblocks
    gsutil cp gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/0006810700.dbin.zst ./testblocks/
fi

rm -rf ./parallel_run/
export PARALLEL="go run ./cmd/testgraph parallel"
export SPARKLE_CMD_PARALLEL_STEP_BLOCKS_STORE_URL=./testblocks
export SEGMENT_1="--start-block 6810753 --stop-block 6810757"
export SEGMENT_2="--start-block 6810758 --stop-block 6810762"

$PARALLEL step -s=1 $SEGMENT_1 --output-path ./parallel_run/step1
$PARALLEL step -s=1 $SEGMENT_2 --output-path ./parallel_run/step1

$PARALLEL step -s=2 $SEGMENT_1 --output-path ./parallel_run/step2 --input-path ./parallel_run/step1
$PARALLEL step -s=2 $SEGMENT_2 --output-path ./parallel_run/step2 --input-path ./parallel_run/step1

$PARALLEL step -s=3 $SEGMENT_1 --output-path ./parallel_run/step3 --input-path ./parallel_run/step2 --flush-entities
$PARALLEL step -s=3 $SEGMENT_2 --output-path ./parallel_run/step3 --input-path ./parallel_run/step2 --flush-entities

$PARALLEL to-csv               --output-path ./parallel_run/csv   --input-path ./parallel_run/step3 --chunk-size 10
