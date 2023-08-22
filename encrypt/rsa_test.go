package encrypt

import "testing"

func Test_RSA(t *testing.T){
	key_bytes, pub_bytes 	:= Generate_RSA(BITS)
	key_len 				:= len(key_bytes)
	if key_len == 0 {
		t.Errorf("private key %d", key_len)
	}
	pub_len 				:= len(pub_bytes)
	if pub_len == 0 {
		t.Errorf("private key %d", pub_len)
	}
	
	if !Verify_RSA(key_bytes, pub_bytes) {
		t.Error("private and public key could not be verified")
	}
	
	msg := "Hello world!"
	
	ciphertext := Encrypt_public(msg, pub_bytes)
	if Decrypt_private(ciphertext, key_bytes) != msg {
		t.Error("rsa encryption failed")
	}
	
	cipher_base64 := Encrypt_public_base64(msg, pub_bytes)
	if Decrypt_private_base64(cipher_base64, key_bytes) != msg {
		t.Error("rsa encryption base64 failed")
	}
	
	signature 	:= Sign(msg, key_bytes)
	if !Verify(msg, signature, pub_bytes) {
		t.Error("signature could not be verified")
	}
	
	sig_base64 	:= Sign_base64(msg, key_bytes)
	if !Verify_base64(msg, sig_base64, pub_bytes) {
		t.Error("signature could not be verified with base64 encoding")
	}
}