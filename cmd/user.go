package cmd

import (
	"fmt"
	"os/user"
)

func Get_user() (*user.User, error){
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("Could not get current user: %w", err)
	}
	return usr, nil
}

func User_allowed(allowed_user string) (bool, error){
	usr, err := user.Lookup(allowed_user)
	if err != nil {
		return false, fmt.Errorf("Could not get allowed user: %w", err)
	}
	curr, err := Get_user()
	if err != nil {
		return false, err
	}
	if *curr != *usr {
		return false, nil
	}
	return true, nil
}