package serv

import "net/http"

//	Set cookie on client
func Set_cookie(w http.ResponseWriter, name, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:			name,
		Value:			value,
		Path:			"/",
		MaxAge:			max_age,
		SameSite:		http.SameSiteLaxMode,	// http.SameSiteStrictMode
		Secure:			true,
		HttpOnly:		true,
	})
}

//	Set session cookie on client
func Set_cookie_session(w http.ResponseWriter, name, value string){
	http.SetCookie(w, &http.Cookie{
		Name:			name,
		Value:			value,
		Path:			"/",
		MaxAge:			0,
		SameSite:		http.SameSiteLaxMode,	// http.SameSiteStrictMode
		Secure:			true,
		HttpOnly:		true,
	})
}

//	Set cookie on client without HttpOnly for javascript access
func Set_cookie_script(w http.ResponseWriter, name, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:			name,
		Value:			value,
		Path:			"/",
		MaxAge:			max_age,
		SameSite:		http.SameSiteLaxMode,	// http.SameSiteStrictMode
		Secure:			true,
		HttpOnly:		false,
	})
}

//	Delete cookie on client
func Delete_cookie(w http.ResponseWriter, name string){
	http.SetCookie(w, &http.Cookie{
		Name:			name,
		Value:			"",
		Path:			"/",
		MaxAge:			-1,
		SameSite:		http.SameSiteLaxMode,	// http.SameSiteStrictMode
		Secure:			true,
	})
}