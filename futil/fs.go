package futil

import (
	"os"
	"io/fs"
)

var (
	CHMOD_WR_OWNER fs.FileMode 	= 0640
)

func Chmod_write_owner(path string){
	err := os.Chmod(path, CHMOD_WR_OWNER)
	if err != nil {
		panic("CHMOD write owner: "+err.Error())
	}
}