#!/bin/bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
SUBGRAPH="pancake/prediction/v1=sgd3@QmcSgkFTp19UW675QHWjjmuYm8ggJJzEaHYr7PMQzGKSt2"
PSQL_DSN="${PSQL_DSN:-"postgresql://postgres:${PG_PASSWORD}@127.0.0.1:5432/graph-node?enable_incremental_sort=off&sslmode=disable"}"
main() {
  pushd "$ROOT" &> /dev/null

  go install -v ./cmd/sparkle

#  sparkle \
#    deploy \
#    --psql-dsn="${PSQL_DSN}"\
#    "$SUBGRAPH"

  sparkle \
    index \
    --psql-dsn="${PSQL_DSN}" \
    --rpc-endpoint=http://localhost:8545 \
    --sf-api-key="${DFUSE_API_KEY}" \
    --sf-endpoint=${DFUSE_SF_ENDPOINT} \
    --flush-with-transaction=true \
    "$SUBGRAPH" \
    $@
  popd &> /dev/null
}

main $@
