package sess

import (
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"github.com/clarkk/go-util/serv"
)

const csrf_token = "csrf_token"

func Verify_CSRF(r *http.Request) bool {
	s := Request(r)
	if s == nil {
		return false
	}
	token := s.CSRF_token()
	return token != "" && token == r.Header.Get("X-CSRF-token")
}

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