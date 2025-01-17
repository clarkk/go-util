package sess

import (
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"github.com/clarkk/go-util/hash"
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
	
	token := hash.SHA256_hex([]byte(s.sess.sid+uuid_string()))
	s.data[csrf_token]		= token
	s.sess.data[csrf_token]	= token
	serv.Set_cookie_script(s.w, csrf_token, token, 0)
}