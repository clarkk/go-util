package securetoken

import (
	"crypto/rand"
	"encoding/base64"
	"math"
)

func Token(length int) (string, error){
	n_bytes := int(math.Ceil(float64(length * 6) / 8))
	
	b := make([]byte, n_bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	
	s := base64.RawURLEncoding.EncodeToString(b)
	return s[:length], nil
}