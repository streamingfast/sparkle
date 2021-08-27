package indexer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/streamingfast/dstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRpcCache(t *testing.T) {
	ctx := context.Background()

	readStore, err := dstore.NewStore("file:///tmp/testrpcread", "json.zst", "zstd", false)
	require.NoError(t, err)
	writeStore, err := dstore.NewStore("file:///tmp/testrpcwrite", "json.zst", "zstd", true)
	require.NoError(t, err)

	c := NewCache(readStore, writeStore, 1000, 11000)
	c.Load(ctx)
	k := c.Key("testtype", 234, "something", "blah:\nblah")
	assert.Equal(t, k, RPCCacheKey("testtype:234:something:blah:\nblah"))

	ent := &fakeEntity{
		A: "a",
		B: 2,
		C: "c",
	}
	c.Set(k, ent)

	loadedJSON, found := c.GetJSON(k)
	assert.True(t, found)

	loaded := &fakeEntity{}
	assert.NotEqual(t, loaded, ent)
	err = json.Unmarshal(loadedJSON, &loaded)
	require.NoError(t, err)
	assert.Equal(t, loaded, ent)

	loadedDirect := &fakeEntity{}
	found = c.Get(k, loadedDirect)
	assert.Equal(t, loadedDirect, ent)

	unloadable := &struct {
		A int
	}{}
	found = c.Get(k, unloadable)
	assert.False(t, found)

	wrongKey := c.Key("wrong", 123)
	_, found = c.GetJSON(wrongKey)
	assert.False(t, found)

	// slice
	ents := []*fakeEntity{
		{"aa", 22, "cc"},
		{"aaa", 222, "ccc"},
	}
	c.Set(k, ents)

	loadedEnts := []*fakeEntity{}
	loadedEntsJSON, found := c.GetJSON(k)
	assert.True(t, found)

	err = json.Unmarshal(loadedEntsJSON, &loadedEnts)
	require.NoError(t, err)

	assert.Equal(t, loadedEnts, ents)

	c.Save(context.Background())

}

type fakeEntity struct {
	A string
	B int
	C interface{}
}
