package encrypt

import (
	"io"
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

func Encrypt_AES256_GCM(msg, passphrase string) ([]byte, error){
	if len(passphrase) != 32 {
		return nil, fmt.Errorf("AES-256 passphrase must be 32 bytes")
	}
	
	gcm, err := gcm_cipher([]byte(passphrase))
	if err != nil {
		return nil, err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("Unable to generate random nonce: %w", err)
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(msg), nil)
	return ciphertext, nil
}

func Encrypt_AES256_GCM_base64(msg, passphrase string) (string, error){
	ciphertext, err := Encrypt_AES256_GCM(msg, passphrase)
	return base64.URLEncoding.EncodeToString(ciphertext), err
}

func Decrypt_AES256_GCM(ciphertext []byte, passphrase string) (string, error){
	gcm, err := gcm_cipher([]byte(passphrase))
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("Unable to decrypt ciphertext")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to decrypt ciphertext")
	}
	
	return string(plaintext), nil
}

func Decrypt_AES256_GCM_base64(ciphertext, passphrase string) (string, error){
	ciphertext_bytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	return Decrypt_AES256_GCM(ciphertext_bytes, passphrase)
}

func gcm_cipher(passphrase []byte) (cipher.AEAD, error){
	c, err := aes.NewCipher(passphrase)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate AES cipher: %w", err)
	}
	
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate GCM: %w", err)
	}
	
	return gcm, nil
}