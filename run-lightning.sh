#!/bin/bash

go install -v ./cmd/sparkle || exit 1

STOREURL=gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1
if test -d ./localblocks; then  
  echo "Using blocks from local store: ./localblocks"
  STOREURL=./localblocks
else
  echo "Fetching blocks from remote store. You should copy them locally to make this faster..., ex:"
  cat <<EOC

######

mkdir ./localblocks
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00005*" ./localblocks/
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00006*" ./localblocks/
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00007*" ./localblocks/
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00008*" ./localblocks/

######

EOC
fi

function step1() {
    INFO=.* sparkle lightning step pancake/exchange/v1 -s 1 --output-path ./step1-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step2() {
    INFO=.* sparkle lightning step pancake/exchange/v1 -s 2 --input-path ./step1-v1 --output-path ./step2-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step3() {
    INFO=.* sparkle lightning step pancake/exchange/v1 -s 3 --input-path ./step2-v1 --output-path ./step3-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step4() {
    INFO=.* sparkle lightning step pancake/exchange/v1 -s 4 --flush-entities --store-snapshot=false --input-path ./step3-v1 --output-path ./step4-v1  --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}

if [ "$1" != "" ] && [ "$1" != 1 ]; then
	echo "SKIPPING STEP 1"
else
	echo "LAUNCHING STEP 1"
	rm -rf ./step1-v1
	
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
fi

if [ "$1" != "" ] && [ "$1" != 2 ]; then
	echo "SKIPPING STEP 2"
else
	echo "LAUNCHING STEP 2"
	rm -rf ./step2-v1
	
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
fi

if [ "$1" != "" ] && [ "$1" != 3 ]; then
	echo "SKIPPING STEP 3"
else
	echo "LAUNCHING STEP 3"
	rm -rf ./step3-v1
	
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
fi

if [ "$1" != "" ] && [ "$1" != 4 ]; then
	echo "SKIPPING STEP 4"
else
	echo "LAUNCHING STEP 4"
	rm -rf ./step4-v1
	
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
fi

if [ "$1" != "" ] && [ "$1" != csv ]; then
	echo "SKIPPING STEP CSV"
else
	echo "Exporting to csv"
	INFO=.* sparkle lightning to-csv pancake/exchange/v1 --input-path ./step4-v1 --output-path ./stepcsv-v1 --chunk-size 1000
fi




