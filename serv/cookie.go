package serv

import (
	"net/http"
)

func Set_cookie(w http.ResponseWriter, name string, value string, max_age int){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		value,
		MaxAge:		max_age,
		Secure:		true,
		HttpOnly:	true,
	})
}

func Delete_cookie(w http.ResponseWriter, name string){
	http.SetCookie(w, &http.Cookie{
		Name:		name,
		Value:		"",
		MaxAge:		-1,
		Secure:		true,
		HttpOnly:	true,
	})
}
