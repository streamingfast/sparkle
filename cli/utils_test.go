package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/test-go/testify/require"
)

func TestParseSubgraphVersionedName(t *testing.T) {
	tests := []struct {
		name                  string
		subgraphVersionedName string
		expectVersion         *VersionedSubgraph
		expectError           bool
	}{
		{
			name:        "subgraph without version",
			expectError: true,
		},
		{
			name:                  "subgraph with version alias",
			subgraphVersionedName: "pancake/exchange-v2@current",
			expectVersion:         &VersionedSubgraph{"pancake/exchange-v2", "current"},
		},
		{
			name:                  "subgraph with version id",
			subgraphVersionedName: "pancake/exchange-v2@1234567890",
			expectVersion:         &VersionedSubgraph{"pancake/exchange-v2", "1234567890"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := parseSubgraphVersionedName(test.subgraphVersionedName)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, v, test.expectVersion)
			}
		})
	}

}
