package encrypt

import "testing"

func Test_RSA(t *testing.T){
	key_bytes, pub_bytes, _	:= Generate_RSA(BITS4096)
	key_len 				:= len(key_bytes)
	if key_len == 0 {
		t.Errorf("private key %d", key_len)
	}
	pub_len 				:= len(pub_bytes)
	if pub_len == 0 {
		t.Errorf("private key %d", pub_len)
	}
	
	if !Verify_RSA(key_bytes, pub_bytes) {
		t.Errorf("private and public key could not be verified")
	}
	
	msg := "Hello world!"
	
	ciphertext, _		:= Encrypt_public(msg, pub_bytes)
	if decrypt_msg, _	:= Decrypt_private(ciphertext, key_bytes); decrypt_msg != msg {
		t.Errorf("rsa encryption failed")
	}
	
	cipher_base64, _	:= Encrypt_public_base64(msg, pub_bytes)
	if decrypt_msg, _	:= Decrypt_private_base64(cipher_base64, key_bytes); decrypt_msg != msg {
		t.Errorf("rsa encryption base64 failed")
	}
	
	signature, _ 		:= Sign(msg, key_bytes)
	if !Verify(msg, signature, pub_bytes) {
		t.Errorf("signature could not be verified")
	}
	
	sig_base64, _		:= Sign_base64(msg, key_bytes)
	if !Verify_base64(msg, sig_base64, pub_bytes) {
		t.Errorf("signature could not be verified with base64 encoding")
	}
}