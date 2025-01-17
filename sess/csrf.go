package sess

import (
	"fmt"
	"net/http"
	"github.com/clarkk/go-util/hash"
	"github.com/clarkk/go-util/serv"
)

const csrf_token = "csrf_token"

func Verify_CSRF(r *http.Request) bool {
	s := Request(r)
	if s == nil {
		return false
	}
	token := s.csrf_token()
	fmt.Println("CSRF:", token, r.Header.Get("X-CSRF-token"))
	return token != "" && token == r.Header.Get("X-CSRF-token")
}

func (s *Session) Generate_CSRF() string {
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	token := hash.SHA256_hex([]byte(s.sess.sid+uuid_string()))
	s.data[csrf_token]		= token
	s.sess.data[csrf_token]	= token
	serv.Set_cookie_script(s.w, csrf_token, token, 0)
	return token
}