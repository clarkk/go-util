package secure_pass

import (
	"fmt"
	"math/big"
	"crypto/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(length int) (string, error){
	pass			:= make([]byte, length)
	charset_length	:= big.NewInt(int64(len(charset)))
	for i := range pass {
		index, err := rand.Int(rand.Reader, charset_length)
		if err != nil {
			return "", fmt.Errorf("Unable to generate password: %v", err)
		}
		pass[i] = charset[index.Int64()]
	}
	return string(pass), nil
}