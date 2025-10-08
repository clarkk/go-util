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

func HMAC_SHA256_hex(key, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}