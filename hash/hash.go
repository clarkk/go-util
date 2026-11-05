package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SHA256_hex(b []byte) string {
	return hex.EncodeToString(SHA256(b))
}

func SHA256(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

func HMAC_SHA256_hex(key, msg []byte) string {
	return hex.EncodeToString(HMAC_SHA256(key, msg))
}

func HMAC_SHA256(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}