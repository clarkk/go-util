package sess

import (
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"github.com/clarkk/go-util/serv"
)

const csrf_token = "csrf_token"

func (s *Session) Generate_CSRF(){
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	hash		:= sha256.Sum256([]byte(s.sess.sid+uuid_string()))
	hash_hex	:= hex.EncodeToString(hash[:])
	s.data[csrf_token]		= hash_hex
	s.sess.data[csrf_token]	= hash_hex
	serv.Set_cookie_script(s.w, csrf_token, hash_hex, 0)
}

func (s *Session) Verify_CSRF(r *http.Request) bool {
	header 	:= r.Header.Get("X-CSRF-token")
	token 	:= s.csrf_token()
	if token == "" {
		return true
	}
	return token == header
}