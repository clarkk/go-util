package hash

import (
	"encoding/hex"
	"crypto/sha256"
)

func SHA256_hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}