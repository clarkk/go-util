package sess

import (
	"net/url"
	"net/http"
	"github.com/clarkk/go-util/hash"
	"github.com/clarkk/go-util/serv"
)

const CSRF_HEADER = "X-CSRF-Token"

var (
	csrf_token		string
	csrf_origin		string
)

func Init_CSRF(token, origin string){
	csrf_token		= token
	csrf_origin		= origin
}

func Verify_CSRF(r *http.Request) bool {
	s := Request(r)
	if s == nil {
		return false
	}
	
	header_csrf := r.Header.Get(CSRF_HEADER)
	if header_csrf == "" {
		return false
	}
	
	token := s.csrf_token()
	if token == "" || token != header_csrf {
		return false
	}
	
	if verify_origin(r.Header.Get("Origin")) {
		return true
	}
	
	if verify_origin(r.Header.Get("Referer")) {
		return true
	}
	
	return false
}

func (s *Session) Generate_CSRF(){
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	serv.Set_cookie_script(s.w, csrf_token, s.generate_CSRF(), 0)
}

func (s *Session) generate_CSRF() (token string){
	token = hash.SHA256_hex([]byte(s.sess.sid+uuid_string()))
	s.data.Csrf_token		= token
	s.sess.data.Csrf_token	= token
	return
}

func verify_origin(header_url string) bool {
	if header_url == "" {
		return false
	}
	parsed_url, err := url.Parse(header_url)
	if err != nil {
		return false
	}
	return csrf_origin == parsed_url.Host
}