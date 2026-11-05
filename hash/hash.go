package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SHA256_hex(b []byte) string {
	return string(SHA256_hex_bytes(b))
}

func SHA256_hex_bytes(b []byte) []byte {
	sum := sha256.Sum256(b)
	out := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(out, sum[:])
	return out
}

func HMAC_SHA256_hex(key, msg []byte) string {
	return string(HMAC_SHA256_hex_bytes(key, msg))
}

func HMAC_SHA256_hex_bytes(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	sum := mac.Sum(nil)
	
	out := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(out, sum)
	return out
}