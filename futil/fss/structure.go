package fss

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"path/filepath"
	"github.com/clarkk/go-util/futil"
)

const (
	FSS_MIN_DIGITS 	= 2
	FSS_SEPARATOR 	= "_"
)

//	Get structured file path from file ID
func Get(file_id int, base_path string, min_digits int) string {
	get, _ := compile(file_id, base_path, min_digits, false, false)
	return get
}

//	Get structured file path from file ID with folder
func Get_folder(file_id int, base_path string, min_digits int) string {
	get, _ := compile(file_id, base_path, min_digits, false, true)
	return get
}

//	Check if structed file path exists from file ID with folder
func Exists_folder(file_id int, base_path string, min_digits int) (bool, error) {
	return futil.Exists(Get_folder(file_id, base_path, min_digits))
}

//	Create structured file path from file ID
func Create(file_id int, base_path string, min_digits int) (string, error){
	return compile(file_id, base_path, min_digits, true, false)
}

//	Create structured file path from file ID with folder
func Create_folder(file_id int, base_path string, min_digits int) (string, error){
	return compile(file_id, base_path, min_digits, true, true)
}

//	Fetch files in structured file path by file ID
func Fetch(file_id int, base_path string, min_digits int) ([]string, error){
	files, err := filepath.Glob(Get(file_id, base_path, min_digits)+"/"+strconv.Itoa(file_id)+FSS_SEPARATOR+"*")
	if err != nil {
		return []string{}, fmt.Errorf("Unable to fetch FSS files: %w", err)
	}
	return files, nil
}

//	Fetch files in structured file path by file ID with folder
func Fetch_folder(file_id int, base_path string, min_digits int) ([]string, error){
	files, err := filepath.Glob(Get_folder(file_id, base_path, min_digits)+"/*")
	if err != nil {
		return []string{}, fmt.Errorf("Unable to fetch FSS files: %w", err)
	}
	return files, nil
}

//	Delete empty directories in structured file path
func Purge(path string) error {
	//	Check if path is a directory
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Unable to access directory: %w", err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("Path is not a directory %s", path)
	}
	
	for true {
		//	Check if directory name is digits
		if _, err := strconv.Atoi(filepath.Base(path)); err != nil {
			break
		}
		
		//	Check if directory is empty
		empty, err := futil.Empty_dir(path)
		if err != nil {
			return fmt.Errorf("Unable to access directory: %w", err)
		}
		if !empty {
			break
		}
		
		//	Delete directory
		if err = os.Remove(path); err != nil {
			return fmt.Errorf("Unable to delete directory %s: %w", path, err)
		}
		
		path = filepath.Dir(path)
	}
	
	return nil
}

//	Compile structured file path from file ID
func compile(file_id int, path string, min_digits int, create, folder bool) (string, error){
	id 		:= strconv.Itoa(file_id)
	path 	= strings.TrimRight(path, "/")
	length	:= len(id)
	
	for i := range length {
		len := length - i
		if len <= min_digits {
			break
		}
		
		digit := string(id[i])
		if digit == "0" {
			continue
		}
		
		//	Zero fill right
		path += "/"+digit+strings.Repeat("0", len-1)
	}
	
	if folder {
		path += "/"+id
	}
	
	if create {
		if err := os.MkdirAll(path, futil.CHMOD_RWX_OWNER); err != nil {
			return "", fmt.Errorf("Unable to create FSS path %s: %w", path, err)
		}
	}
	
	return path, nil
}