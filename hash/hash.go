package hash

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

func SHA1_hex(b []byte) string {
	sum := sha1.Sum1(b)
	return hex.EncodeToString(sum[:])
}

func SHA256_hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}