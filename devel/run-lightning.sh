#!/bin/bash

go install -v ./cmd/sparke || exit 1

export SPARKLE_LIGHTNING_STEP_SF_API_KEY=$DFUSE_KEY
# After:
# gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00005*" ./myblocks/
# gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00006*" ./myblocks/
# gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00007*" ./myblocks/
# gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00008*" ./myblocks/
export SPARKLE_LIGHTNING_STEP_CMD_BLOCKS_STORE_URL=./myblocks

echo "LAUNCHING STEP 1"

function step1() {
    INFO=.* sparke lightning step -s 1 --output-path ./step1-v1 --start-block $1 --stop-block $2 &
}
function step2() {
    INFO=.* sparke lightning step -s 2 --input-path ./step1-v1 --output-path ./step2-v1 --start-block $1 --stop-block $2 &
}
function step3() {
    INFO=.* sparke lightning step -s 3 --input-path ./step2-v1 --output-path ./step3-v1 --start-block $1 --stop-block $2 &
}
function step4() {
    INFO=.* sparke lightning step -s 4 --flush-entities --store-snapshot=false --input-path ./step3-v1 --output-path ./step4-v1  --start-block $1 --stop-block $2  &
}

step1 580000 600000
step1 600000 620000
step1 620000 640000
step1 640000 660000
step1 660000 680000
step1 680000 700000
step1 700000 720000

for job in `jobs -p`; do
    echo "Waiting on $job"
    wait $job
done

echo "LAUNCHING STEP 2"

step2 580000 600000
step2 600000 620000
step2 620000 640000
step2 640000 660000
step2 660000 680000
step2 680000 700000
step2 700000 720000

for job in `jobs -p`; do
    echo "Waiting on $job"
    wait $job
done

echo "LAUNCHING STEP 3"

step3 580000 600000
step3 600000 620000
step3 620000 640000
step3 640000 660000
step3 660000 680000
step3 680000 700000
step3 700000 720000

for job in `jobs -p`; do
    echo "Waiting on $job"
    wait $job
done

echo "LAUNCHING STEP 4"

step4 580000 600000
step4 600000 620000
step4 620000 640000
step4 640000 660000
step4 660000 680000
step4 680000 700000
step4 700000 720000

for job in `jobs -p`; do
    echo "Waiting on $job"
    wait $job
done

echo "Exporting to csv"
INFO=.* sparkle lightning to-csv --input-path ./step4-v1 --output-path ./exported




