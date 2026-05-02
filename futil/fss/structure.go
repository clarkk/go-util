package fss

import (
	"os"
	"io/fs"
	"fmt"
	"strings"
	"strconv"
	"path/filepath"
	"github.com/clarkk/go-util/futil"
)

const (
	MIN_DIGITS 	= 2
	SEPARATOR 	= "_"
)

type Path struct {
	file_id		uint64
	base_path	string
	path		string
}

func New(file_id uint64, base_path string, min_digits int) *Path {
	base_path = filepath.Clean(base_path)
	return &Path{
		file_id:	file_id,
		base_path:	base_path,
		path:		compile(file_id, base_path, min_digits),
	}
}

//	Get path
func (p *Path) Get() string {
	return p.path
}

//	Check if path exists
func (p *Path) Exists() (bool, error){
	return futil.Exists(p.path)
}

//	Create path
func (p *Path) Create() (string, error){
	if err := os.MkdirAll(p.path, futil.CHMOD_RWX_OWNER); err != nil {
		return "", fmt.Errorf("Unable to create FSS path %s: %w", p.path, err)
	}
	return p.path, nil
}

//	Write file
func (p *Path) Write(suffix_name string, data []byte, mode fs.FileMode) error {
	file := p.file_prefix()+suffix_name
	if err := os.WriteFile(file, data, mode); err != nil {
		return fmt.Errorf("Unable to write FSS file %s: %w", file, err)
	}
	return nil
}

//	Fetch files by ID + separator
func (p *Path) Fetch() ([]string, error){
	files, err := filepath.Glob(p.file_prefix()+"*")
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch FSS files: %w", err)
	}
	return files, nil
}

//	Delete files by ID + separator
func (p *Path) Clear(purge bool) error {
	files, err := p.Fetch()
	if err != nil {
		return err
	}
	if len(files) > 0 {
		if err = futil.Delete(files); err != nil {
			return err
		}
	}
	if purge {
		return p.Purge()
	}
	return nil
}

//	Delete empty directories in path
func (p *Path) Purge() error {
	//	Check if path is a directory
	stat, err := os.Stat(p.path)
	if err != nil {
		return fmt.Errorf("Unable to access directory: %w", err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("Path is not a directory %s", p.path)
	}
	
	path := p.path
	for {
		if path == p.base_path {
			break
		}
		
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

func (p *Path) file_prefix() string {
	return p.path+"/"+strconv.FormatUint(p.file_id, 10)+SEPARATOR
}

func compile(file_id uint64, base_path string, min_digits int) string {
	id		:= strconv.FormatUint(file_id, 10)
	length	:= len(id)
	
	var sb strings.Builder
	sb.WriteString(base_path)
	
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
	
	return sb.String()
}