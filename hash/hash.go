package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SHA256_hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func HMAC_SHA256(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}