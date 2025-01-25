package hash_pass

import (
	"fmt"
	"bytes"
	"strings"
	"strconv"
	"runtime"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
)

const (
	time uint32			= 1
	memory uint32		= 1024 * 64
	salt_bytes uint32	= 32
	hash_bytes uint32	= 128
)

func Create(password string) (string, error){
	salt, err := random_salt(salt_bytes)
	if err != nil {
		return "", err
	}
	hash := generate_hash(password, salt, time, memory, hash_bytes)
	return fmt.Sprintf("%d:%d:%s:%s",
		time,
		memory,
		base64_encode(salt),
		base64_encode(hash)), nil
}

func Compare(password, hash string) (bool, error){
	s := strings.Split(hash, ":")
	time, err := strconv.Atoi(s[0])
	if err != nil {
		return false, err
	}
	memory, err := strconv.Atoi(s[1])
	if err != nil {
		return false, err
	}
	salt, err := base64_decode(s[2])
	if err != nil {
		return false, err
	}
	hash_compare, err := base64_decode(s[3])
	if err != nil {
		return false, err
	}
	if !bytes.Equal(hash_compare, generate_hash(password, salt, uint32(time), uint32(memory), uint32(len(hash_compare)))) {
		return false, nil
	}
	return true, nil
}

func generate_hash(password string, salt []byte, time, memory uint32, hash_bytes uint32) []byte {
	return argon2.IDKey([]byte(password), salt, time, memory, uint8(runtime.NumCPU()), hash_bytes)
}

func random_salt(length uint32) ([]byte, error){
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func base64_encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func base64_decode(s string) ([]byte, error){
	return base64.StdEncoding.DecodeString(s)
}