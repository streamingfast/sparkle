package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateID(t *testing.T) {
	id := generateID()
	assert.Equal(t, len(id), 32)
}
