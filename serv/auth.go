package serv

import (
	"time"
	"net/http"
	"crypto/subtle"
)

func Auth_basic(r *http.Request, auth_user, auth_pass string) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	
	user_match := subtle.ConstantTimeCompare([]byte(user), []byte(auth_user)) == 1
	pass_match := subtle.ConstantTimeCompare([]byte(pass), []byte(auth_pass)) == 1
	
	if !user_match || !pass_match {
		time.Sleep(2 * time.Second)
		return false
	}
	
	return true
}