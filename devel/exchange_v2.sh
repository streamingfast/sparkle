#!/bin/bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
SUBGRAPH="pancake/exchange/v2=sgd2@QmdMEBfoi8b5Q2CLeNo3UyLW8ow5PofzdLhwFY26QRrDfw"

main() {
  pushd "$ROOT" &> /dev/null

  go install ./cmd/sparkle

  sparkle \
    deploy \
    --psql-dsn="postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable" \
    "$SUBGRAPH"

  sparkle \
    index \
    --psql-dsn="postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable" \
    --rpc-endpoint=http://localhost:8545 \
    --sf-api-key="${DFUSE_API_KEY}" \
    --sf-endpoint=${DFUSE_SF_ENDPOINT} \
    --flush-with-transaction=true \
    "$SUBGRAPH" \
    $@
  popd &> /dev/null
}

main $@
