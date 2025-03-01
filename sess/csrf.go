package sess

import (
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
	
	cookie, err := r.Cookie(csrf_token)
	if err != nil {
		return false
	}
	
	token := s.csrf_token()
	return token != "" && token == cookie.Value
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