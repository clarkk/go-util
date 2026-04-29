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
func Get(file_id uint64, base_path string, min_digits int) string {
	get, _ := compile(file_id, base_path, min_digits, false, false)
	return get
}

//	Get structured file path from file ID with directory
func Get_dir(file_id uint64, base_path string, min_digits int) string {
	get, _ := compile(file_id, base_path, min_digits, false, true)
	return get
}

//	Check if structed file path exists from file ID with directory
func Exists_dir(file_id uint64, base_path string, min_digits int) (bool, error) {
	return futil.Exists(Get_dir(file_id, base_path, min_digits))
}

//	Create structured file path from file ID
func Create(file_id uint64, base_path string, min_digits int) (string, error){
	return compile(file_id, base_path, min_digits, true, false)
}

//	Create structured file path from file ID with directory
func Create_dir(file_id uint64, base_path string, min_digits int) (string, error){
	return compile(file_id, base_path, min_digits, true, true)
}

//	Fetch files in structured file path by file ID
func Fetch(file_id uint64, base_path string, min_digits int) ([]string, error){
	files, err := filepath.Glob(Get(file_id, base_path, min_digits)+"/"+strconv.FormatUint(file_id, 10)+FSS_SEPARATOR+"*")
	if err != nil {
		return []string{}, fmt.Errorf("Unable to fetch FSS files: %w", err)
	}
	return files, nil
}

//	Fetch files in structured file path by file ID with directory
func Fetch_dir(file_id uint64, base_path string, min_digits int) ([]string, error){
	files, err := filepath.Glob(Get_dir(file_id, base_path, min_digits)+"/*")
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
		if _, err := strconv.ParseUint(filepath.Base(path), 10, 64); err != nil {
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
func compile(file_id uint64, base_path string, min_digits int, create, directory bool) (string, error){
	id		:= strconv.FormatUint(file_id, 10)
	length	:= len(id)
	
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(base_path, "/"))
	
	for i := range length {
		remain := length - i
		if remain <= min_digits {
			break
		}
		
		digit := id[i]
		if digit == '0' {
			continue
		}
		
		//	Zero fill right
		sb.WriteByte('/')
		sb.WriteByte(digit)
		for range remain - 1 {
			sb.WriteByte('0')
		}
	}
	
	if directory {
		sb.WriteByte('/')
		sb.WriteString(id)
	}
	
	path := sb.String()
	
	if create {
		if err := os.MkdirAll(path, futil.CHMOD_RWX_OWNER); err != nil {
			return "", fmt.Errorf("Unable to create FSS path %s: %w", path, err)
		}
	}
	
	return path, nil
}