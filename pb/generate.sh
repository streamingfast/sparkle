#!/bin/bash
# Copyright 2019 dfuse Platform Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

# Protobuf definitions
PROTO=${1:-"$ROOT/../proto"}
PROTO_ETHEREUM=${2:-"$ROOT/../proto-ethereum"}

function main() {
  current_dir="`pwd`"
  trap "cd \"$current_dir\"" EXIT
  pushd "$ROOT/pb" &> /dev/null

 generate "dfuse/ethereum/codec/v1/codec.proto"
 generate "dfuse/ethereum/tokenmeta/v1/tokenmeta.proto"

  echo "generate.sh - `date` - `whoami`" > $ROOT/pb/last_generate.txt
  echo "streamingfast/proto revision: `GIT_DIR=$PROTO/.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
  echo "streamingfast/proto-ethereum revision: `GIT_DIR=$PROTO_ETHEREUM/.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
}

# usage:
# - generate <protoPath>
# - generate <protoBasePath/> [<file.proto> ...]
function generate() {
    base=""
    if [[ "$#" -gt 1 ]]; then
      base="$1"; shift
    fi

    for file in "$@"; do
      protoc -I$PROTO -I$PROTO_ETHEREUM $base$file --go_out=plugins=grpc,paths=source_relative:.
    done
}

main "$@"

#sed -i s@dfuse-io/dfuse-ethereum/pb/dfuse/ethereum/codec/v1@dfuse-io/dtrader/pb/dfuse/ethereum/codec/v1@ ./dfuse/ethereum/trxstatetracker/v1/trxstatetracker.pb.go
