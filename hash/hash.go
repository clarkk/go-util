package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Sum256 [sha256.Size]byte

func SHA256(b []byte) Sum256 {
	return Sum256(sha256.Sum256(b))
}

func HMAC_SHA256(key, msg []byte) Sum256 {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	
	var sum Sum256
	mac.Sum(sum[:0])
	return sum
}

func (s Sum256) Bytes() []byte {
	return s[:]
}

func (s Sum256) Hex() string {
	return hex.EncodeToString(s[:])
}

func (s Sum256) Hex_bytes() []byte {
	out := make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(out, s[:])
	return out
}