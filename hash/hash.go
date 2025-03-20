package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256_hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}