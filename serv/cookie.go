package serv

import "net/http"

//	Set cookie on client
func Set_cookie(w http.ResponseWriter, name, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		value,
		MaxAge:		max_age,
		Path:		"/",
		SameSite:	http.SameSiteStrictMode,//http.SameSiteLaxMode,
		Secure:		true,
		HttpOnly:	true,
	})
}

//	Set cookie on client without HttpOnly
func Set_cookie_script(w http.ResponseWriter, name, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		value,
		MaxAge:		max_age,
		Path:		"/",
		SameSite:	http.SameSiteStrictMode,//http.SameSiteLaxMode,
		Secure:		true,
		HttpOnly:	false,
	})
}

//	Delete cookie on client
func Delete_cookie(w http.ResponseWriter, name string){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		"",
		MaxAge:		-1,
	})
}