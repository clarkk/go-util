package serv

import "net/http"

//	Set cookie on client
func Set_cookie(w http.ResponseWriter, name string, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		value,
		MaxAge:		max_age,
		Path:		"/",
		SameSite:	http.SameSiteLaxMode,
		Secure:		true,
		HttpOnly:	true,
	})
}

//	Delete cookie on client
func Delete_cookie(w http.ResponseWriter, name string){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		"",
		MaxAge:		-1,
		Path:		"/",
		SameSite:	http.SameSiteLaxMode,
		Secure:		true,
		HttpOnly:	true,
	})
}