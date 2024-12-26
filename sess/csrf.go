package sess

import (
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"github.com/google/uuid"
	"github.com/clarkk/go-util/serv"
)

const csrf_token = "csrf_token"

func (s *Session) Generate_CSRF(){
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	hash		:= sha256.Sum256([]byte(s.sid+uuid.NewString()))
	hash_hex	:= hex.EncodeToString(hash[:])
	s.data[csrf_token] = hash_hex
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