package hash_pass

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const reset_token_length = 32

func Reset_token() (token, token_hash string, err error){
	b := make([]byte, reset_token_length)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	token		= hex.EncodeToString(b)
	token_hash	= Reset_token_hash(token)
	return
}

func Reset_token_hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}