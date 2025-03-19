package encrypt

import (
	"fmt"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/sha512"
	"encoding/pem"
	"encoding/base64"
)

const BITS4096 = 4096

func Generate_RSA(bits int) ([]byte, []byte, error){
	//	Generate private key
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("Unable to generate private key: %v", err)
	}
	
	//	Encode private key to PKCS#1 PEM
	key_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:	"RSA PRIVATE KEY",
			Bytes:	x509.MarshalPKCS1PrivateKey(key),
		},
	)
	
	//	Encode public key to PKCS#1 PEM
	pub_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:	"RSA PUBLIC KEY",
			Bytes:	x509.MarshalPKCS1PublicKey(&key.PublicKey),
		},
	)
	
	return key_pem, pub_pem, nil
}

func Verify_RSA(private, public []byte) bool {
	key := decode_private_pem(private)
	pub := decode_public_pem(public)
	return key.PublicKey.Equal(pub)
}

func Encrypt_public(msg string, public []byte) ([]byte, error){
	var label []byte
	ciphertext, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, decode_public_pem(public), []byte(msg), label)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to encrypt: %v", err)
	}
	return ciphertext, nil
}

func Encrypt_public_base64(msg string, public []byte) (string, error){
	ciphertext, err := Encrypt_public(msg, public)
	return base64.URLEncoding.EncodeToString(ciphertext), err
}

func Decrypt_private(ciphertext, private []byte) (string, error){
	var label []byte
	text, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, decode_private_pem(private), ciphertext, label)
	if err != nil {
		return "", fmt.Errorf("Unable to decrypt: %v", err)
	}
	return string(text), nil
}

func Decrypt_private_base64(ciphertext string, private []byte) (string, error){
	ciphertext_bytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	return Decrypt_private(ciphertext_bytes, private)
}

func Sign(msg string, private []byte) ([]byte, error){
	signature, err := rsa.SignPKCS1v15(rand.Reader, decode_private_pem(private), crypto.SHA512, digest(msg))
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to sign: %v", err)
	}
	return signature, nil
}

func Sign_base64(msg string, private []byte) (string, error){
	signature, err := Sign(msg, private)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(signature), nil
}

func Verify(msg string, signature, public []byte) bool {
	err := rsa.VerifyPKCS1v15(decode_public_pem(public), crypto.SHA512, digest(msg), signature)
	if err != nil {
		return false
	}
	return true
}

func Verify_base64(msg, signature string, public []byte) bool {
	signature_bytes, _ := base64.URLEncoding.DecodeString(signature)
	return Verify(msg, signature_bytes, public)
}

func decode_public_pem(public []byte) *rsa.PublicKey {
	block, _ := pem.Decode(public)
	pub, _ := x509.ParsePKCS1PublicKey(block.Bytes)
	return pub
}

func decode_private_pem(private []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(private)
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	return key
}

func digest(msg string) []byte {
	digest := sha512.Sum512([]byte(msg))
	return digest[:]
}