package futil

import (
	"os"
	"io"
	"fmt"
	"io/fs"
	"errors"
)

const (
	CHMOD_RW_OWNER fs.FileMode 	= 0644
	CHMOD_RWX_OWNER fs.FileMode = 0744
)

//	Check if file/directory exists
func Exists(path string) (bool, error){
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//	Check if directory is empty
func Empty_dir(path string) (bool, error){
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err != nil && errors.Is(err, io.EOF) {
		return true, nil
	}
	return false, err
}

//	Delete slice of files
func Delete(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("Unable to delete file %s: %w", file, err)
		}
	}
	return nil
}

func Chmod_rw_owner(path string) error {
	if err := os.Chmod(path, CHMOD_RW_OWNER); err != nil {
		return err
	}
	return nil
}

func Chmod_rwx_owner(path string) error {
	if err := os.Chmod(path, CHMOD_RWX_OWNER); err != nil {
		return err
	}
	return nil
}