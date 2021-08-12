#!/bin/bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
SUBGRAPH="pancake/exchange/v1=sgd1@QmekP583qkqbkhx54kyZC3pviSqoAdjcx94sme1mZ9shv1"

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
