package deployment

import (
	"crypto/rand"
	"encoding/hex"
)

func generateID() string {
	data := make([]byte, 16)
	rand.Read(data)
	return hex.EncodeToString(data)
}
