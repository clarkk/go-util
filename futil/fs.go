package futil

import (
	"os"
)

func Chmod_write_owner(path string){
	err := os.Chmod(path, 0640)
	if err != nil {
		panic(err)
	}
}