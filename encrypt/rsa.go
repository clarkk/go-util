package encrypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/sha512"
	"encoding/pem"
	"encoding/base64"
)

const (
	BITS4096 	= 4096
)

func Generate_RSA(bits int) ([]byte, []byte){
	//	Generate private key
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic("Generate RSA keys: "+err.Error())
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
	
	return key_pem, pub_pem
}

func Verify_RSA(private []byte, public []byte) bool{
	key := decode_private_pem(private)
	pub := decode_public_pem(public)
	return key.PublicKey.Equal(pub)
}

func Encrypt_public(msg string, public []byte) []byte{
	var label []byte
	ciphertext, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, decode_public_pem(public), []byte(msg), label)
	if err != nil {
		panic("Encrypt public: "+err.Error())
	}
	return ciphertext
}

func Encrypt_public_base64(msg string, public []byte) string{
	return base64.URLEncoding.EncodeToString(Encrypt_public(msg, public))
}

func Decrypt_private(ciphertext []byte, private []byte) string{
	var label []byte
	text, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, decode_private_pem(private), ciphertext, label)
	if err != nil {
		panic("Decrypt private: "+err.Error())
	}
	return string(text)
}

func Decrypt_private_base64(ciphertext string, private []byte) string{
	ciphertext_bytes, _ := base64.URLEncoding.DecodeString(ciphertext)
	return Decrypt_private(ciphertext_bytes, private)
}

func Sign(msg string, private []byte) []byte{
	signature, err := rsa.SignPKCS1v15(rand.Reader, decode_private_pem(private), crypto.SHA512, digest(msg))
	if err != nil {
		panic("Sign message: "+err.Error())
	}
	return signature
}

func Sign_base64(msg string, private []byte) string{
	return base64.URLEncoding.EncodeToString(Sign(msg, private))
}

func Verify(msg string, signature []byte, public []byte) bool{
	err := rsa.VerifyPKCS1v15(decode_public_pem(public), crypto.SHA512, digest(msg), signature)
	if err != nil {
		return false
	}
	return true
}

func Verify_base64(msg string, signature string, public []byte) bool{
	signature_bytes, _ := base64.URLEncoding.DecodeString(signature)
	return Verify(msg, signature_bytes, public)
}

func decode_public_pem(public []byte) *rsa.PublicKey{
	block, _ := pem.Decode(public)
	pub, _ := x509.ParsePKCS1PublicKey(block.Bytes)
	return pub
}

func decode_private_pem(private []byte) *rsa.PrivateKey{
	block, _ := pem.Decode(private)
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	return key
}

func digest(msg string) []byte{
	digest := sha512.Sum512([]byte(msg))
	return digest[:]
}