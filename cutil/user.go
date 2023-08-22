package cutil

import (
	"log"
	"os/user"
)

func Get_user() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Could not get current user")
	}
	return usr
}

func User_allowed(allowed_user string){
	usr, err := user.Lookup(allowed_user)
	if err != nil {
		log.Fatal("Could not get allowed user")
	}
	curr := Get_user()
	if *curr != *usr {
		log.Fatal("User is not "+allowed_user)
	}
}