package sess

import (
	"fmt"
	"net/http"
	"github.com/clarkk/go-util/hash"
	"github.com/clarkk/go-util/serv"
)

var (
	csrf_header		string
	csrf_token		string
	csrf_referer	string
)

func Init_CSRF(header, token, referer string){
	csrf_header		= header
	csrf_token		= token
	csrf_referer	= referer
}

func Verify_CSRF(r *http.Request) bool {
	s := Request(r)
	if s == nil {
		return false
	}
	
	header_csrf	:= r.Header.Get(csrf_header)
	token		:= s.csrf_token()
	referer		:= r.Header.Get("Referer")
	fmt.Println("header:", header_csrf, "session:", token, "referer:", referer, csrf_referer)
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