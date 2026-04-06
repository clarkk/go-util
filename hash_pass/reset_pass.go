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
	
	token = hex.EncodeToString(b)
	
	hash := sha256.Sum256([]byte(token))
	token_hash = hex.EncodeToString(hash[:])
	
	return
}