package hash

import (
	"encoding/hex"
	"crypto/sha256"
)

func SHA256_hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}