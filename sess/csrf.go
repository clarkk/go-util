package sess

import (
	"fmt"
	"net/http"
	"github.com/clarkk/go-util/hash"
	"github.com/clarkk/go-util/serv"
)

const (
	header_csrf_token	= "X-CSRF-Token"
	csrf_token			= "csrf_token"
)

func Verify_CSRF(r *http.Request) bool {
	s := Request(r)
	if s == nil {
		return false
	}
	
	header_csrf	:= r.Header.Get(header_csrf_token)
	token		:= s.csrf_token()
	fmt.Println("header:", header_csrf, "session:", token)
	return token != "" && token == header_csrf
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