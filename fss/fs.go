package fss

import (
	"strings"
	"os"
	//"io"
	"path/filepath"
	"strconv"
)

const (
	MIN_DIGITS 	= 2
	DIR_PERM 	= 0640
	SEPARATOR 	= "_"
)

//	Get structured file path from file ID
func Get(file_id int, base_path string, min_digits int) string{
	return compile(file_id, base_path, min_digits, false, false)
}

//	Get structured file path from file ID with folder
func Get_folder(file_id int, base_path string, min_digits int) string{
	return compile(file_id, base_path, min_digits, false, true)
}

//	Create structured file path from file ID
func Create(file_id int, base_path string, min_digits int) string{
	return compile(file_id, base_path, min_digits, true, false)
}

//	Create structured file path from file ID with folder
func Create_folder(file_id int, base_path string, min_digits int) string{
	return compile(file_id, base_path, min_digits, true, true)
}

//	Fetch files in structured file path by file ID
func Fetch(file_id int, base_path string, min_digits int) []string{
	files, err := filepath.Glob(Get(file_id, base_path, min_digits)+"/"+strconv.Itoa(file_id)+SEPARATOR+"*")
	if err != nil {
		panic("Could not fetch files: "+err.Error())
	}
	return files
}

//	Fetch files in structured file path by file ID with folder
func Fetch_folder(file_id int, base_path string, min_digits int) []string{
	files, err := filepath.Glob(Get_folder(file_id, base_path, min_digits)+"/*")
	if err != nil {
		panic("Could not fetch files: "+err.Error())
	}
	return files
}

//	Delete files
func Delete(files []string){
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			panic("Could not delete file: "+file+" "+err.Error())
		}
	}
}

//	Delete empty directories in structured file path
/*func Purge(path string){
	//	Check if path is a directory
	stat, err := os.Stat(path)
	if err != nil {
		panic("Could not access directory: "+path+" "+err.Error())
	}
	if !stat.IsDir() {
		panic("Path is not a directory: "+path)
	}
	
	for true {
		//	Check if directory name is digits
		if _, err := strconv.Atoi(filepath.Base(path)); err != nil {
			break
		}
		
		//	Check if directory is empty
		empty, err := empty_dir(path)
		if err != nil {
			panic("Could not access directory: "+path+" "+err.Error())
		}
		if !empty {
			break
		}
		
		//	Delete directory
		err = os.Remove(path)
		if err != nil {
			panic("Could not delete directory: "+path+" "+err.Error())
		}
		
		path = filepath.Dir(path)
	}
}*/

//	Compile structured file path from file ID
func compile(file_id int, path string, min_digits int, create bool, folder bool) string{
	id 		:= strconv.Itoa(file_id)
	path 	= strings.TrimRight(path, "/")
	length	:= len(id)
	
	for i := 0; i < length; i++ {
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
		err := os.MkdirAll(path, DIR_PERM)
		if err != nil {
			panic("Could not create path: "+path+" "+err.Error())
		}
	}
	
	return path
}

//	Check if directory is empty
/*func empty_dir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}*/