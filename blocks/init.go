// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blocks

import (
	"time"

	"github.com/streamingfast/bstream"
)

func init() {
	bstream.GetBlockReaderFactory = bstream.BlockReaderFactoryFunc(blockReaderFactory)
	bstream.GetBlockDecoder = bstream.BlockDecoderFunc(BlockDecoder)
	bstream.GetProtocolFirstStreamableBlock = 1
	//bstream.GetProtocolGenesisBlock = 0
	bstream.GetBlockWriterHeaderLen = 10
	bstream.GetBlockPayloadSetter = bstream.MemoryBlockPayloadSetter
	bstream.GetMemoizeMaxAge = 20 * time.Second
}
