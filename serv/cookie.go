package serv

import "net/http"

//	Set cookie on client
func Set_cookie(w http.ResponseWriter, name, value string, max_age int){
	set_cookie(w, name, value, max_age, true)
}

//	Set session cookie on client
func Set_cookie_session(w http.ResponseWriter, name, value string){
	set_cookie(w, name, value, 0, true)
}

//	Set cookie on client without HttpOnly for javascript access
func Set_cookie_script(w http.ResponseWriter, name, value string, max_age int){
	set_cookie(w, name, value, max_age, false)
}

//	Delete cookie on client
func Delete_cookie(w http.ResponseWriter, name string){
	set_cookie(w, name, "", -1, true)
}

//	Delete cookie on client
func Delete_cookie_script(w http.ResponseWriter, name string){
	set_cookie(w, name, "", -1, false)
}

func set_cookie(w http.ResponseWriter, name, value string, max_age int, http_only bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   max_age,
		Secure:   true,
		HttpOnly: http_only,
		SameSite: http.SameSiteLaxMode,
	})
}