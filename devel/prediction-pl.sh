#!/bin/bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
STOREURL=gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1
if test -d ./localblocks; then
  echo "Using blocks from local store: ./localblocks"
    STOREURL=./localblocks
  else
    echo "Fetching blocks from remote store. You should copy them locally to make this faster..., ex:"
    cat <<EOC
######

mkdir ./localblocks
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/00069519*" ./localblocks/
gsutil -m cp "gs://dfuseio-global-blocks-us/eth-bsc-mainnet/v1/0006952*" ./localblocks/

######
EOC
fi

function step1() {
    INFO=.* sparkle lightning step pancake/prediction/v1 -s 1 --output-path ./step1-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step2() {
    INFO=.* sparkle lightning step pancake/prediction/v1 -s 2 --input-path ./step1-v1 --output-path ./step2-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step3() {
    INFO=.* sparkle lightning step pancake/prediction/v1 -s 3 --input-path ./step2-v1 --output-path ./step3-v1 --start-block $1 --stop-block $2 --blocks-store-url $STOREURL &
}
function step4() {
    INFO=.* sparkle lightning step pancake/prediction/v1 -s 4 --flush-entities --store-snapshot=false --input-path ./step3-v1 --output-path ./step4-v1  --start-block $1 --stop-block $2  --blocks-store-url $STOREURL  &
}


main() {
  pushd "$ROOT" &> /dev/null
    go install -v ./cmd/sparkle || exit 1

    if [ "$1" != "" ] && [ "$1" != 1 ]; then
      echo "SKIPPING STEP 1"
    else
      echo "LAUNCHING STEP 1"
      rm -rf ./step1-v1

      step1 6951900 6951999
      step1 6952000 6952099
      step1 6952100 6952199
      step1 6952200 6952299
      step1 6952300 6952399
      step1 6952400 6952499

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

      step2 6951900 6951999
      step2 6952000 6952099
      step2 6952100 6952199
      step2 6952200 6952299
      step2 6952300 6952399
      step2 6952400 6952499

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

      step3 6951900 6951999
      step3 6952000 6952099
      step3 6952100 6952199
      step3 6952200 6952299
      step3 6952300 6952399
      step3 6952400 6952499

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

      step4 6951900 6951999
      step4 6952000 6952099
      step4 6952100 6952199
      step4 6952200 6952299
      step4 6952300 6952399
      step4 6952400 6952499

      for job in `jobs -p`; do
          echo "Waiting on $job"
          wait $job
      done
    fi

    if [ "$1" != "" ] && [ "$1" != csv ]; then
      echo "SKIPPING STEP CSV"
    else
      echo "Exporting to csv"
#      INFO=.* sparkle lightning to-csv pancake/prediction/v1 --input-path ./step4-v1 --output-path ./stepcsv-v1 --chunk-size 1000
    fi
  popd &> /dev/null
}

main $@



