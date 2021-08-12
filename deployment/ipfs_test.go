package deployment

import (
	"os"
	"testing"
)

func TestUploadManifestToIPFS(t *testing.T) {
	t.Skip("should work off of generate code")
	ipfs, ok := os.LookupEnv("IPFS_ENDPOINT")
	if ipfs == "" || !ok {
		t.Skip("IPFS_ENDPOINT not defined or is empty. skipping test")
	}

	//manifestPath := "/home/colin/go/src/github.com/streamingfast/sparkles/deploy/test_data/subgraph.yaml"
	//hash, err := UploadManifestToIPFS(manifestPath, ipfs)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//t.Logf("uploaded compiled manifest to %s", hash)
}
