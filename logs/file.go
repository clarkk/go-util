package logs

import (
	"os"
	"io"
	"fmt"
	"log"
	"sync"
	"sort"
	"strconv"
	"strings"
	"path/filepath"
	"compress/gzip"
)

const (
	perm_file 	= 0644
	perm_dir 	= 0744
)

type rotate_writer struct {
	lock		sync.Mutex
	f 			*os.File
	file		string
	max_size 	int64
}

func New(file string, max_size_kb int64) (*log.Logger, error){
	w := &rotate_writer{
		file:		file,
		max_size:	max_size_kb * 1024,
	}
	if err := w.create(); err != nil {
		return nil, err
	}
	return log.New(w, "", log.LstdFlags), nil
}

func (w *rotate_writer) Write(b []byte) (int, error){
	w.lock.Lock()
	defer w.lock.Unlock()
	
	finfo, err := os.Stat(w.file)
	if err != nil {
		log.Printf("Log: Unable to get file stat %s: %v", w.file, err)
	} else if finfo.Size() > w.max_size {
		w.rotate()
	}
	return w.f.Write(b)
}

func (w *rotate_writer) rotate(){
	f, err := os.Open(w.file)
	if err != nil {
		log.Printf("Log: Unable to open file (rotation) %s: %v", w.file, err)
		return
	}
	defer f.Close()
	
	dir := w.file+".d"
	if err := os.MkdirAll(dir, perm_dir); err != nil {
		log.Printf("Log: Unable to create gzip directory (rotation) %s: %v", dir, err)
		return
	}
	
	gzip_file, ok := w.compile_gzip_filename(dir)
	if !ok {
		return
	}
	
	if !write_gzip(f, gzip_file) {
		return
	}
	
	//	Truncate log file
	w.f.Truncate(0)
	w.f.Seek(0,0)
}

func (w *rotate_writer) compile_gzip_filename(dir string) (string, bool){
	entries, err := filepath.Glob(dir+"/*.gz")
	if err != nil {
		log.Printf("Log: Unable to read gzip directory files (rotation) %s: %v", dir, err)
		return "", false
	}
	length := len(entries)
	if length == 0 {
		return w.gzip_filename(dir, 1), true
	}
	list := make([]int, length)
	for k, file := range entries {
		file = file[:len(file) - 3]
		pos := strings.LastIndex(file, ".") + 1
		i, err := strconv.Atoi(file[pos:])
		if err != nil {
			log.Printf("Log: Unable to compile gzip file name (rotation) %s: %v", dir, err)
			return "", false
		}
		list[k] = i
	}
	sort.Ints(list)
	next := list[len(list) - 1] + 1
	return w.gzip_filename(dir, next), true
}

func (w *rotate_writer) create() error {
	var err error
	w.f, err = os.OpenFile(w.file, os.O_CREATE | os.O_WRONLY | os.O_APPEND, perm_file)
	return err
}

func (w *rotate_writer) gzip_filename(dir string, i int) string {
	return fmt.Sprintf("%s/%s.%d.gz", dir, filepath.Base(w.file), i)
}

func write_gzip(src *os.File, gzip_file string) bool {
	f, err := os.Create(gzip_file)
	if err != nil {
		log.Printf("Log: Unable to create gzip file (rotation) %s: %v", gzip_file, err)
		return false
	}
	defer f.Close()
	w, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		log.Printf("Log: Unable to write gzip file (rotation) %s: %v", gzip_file, err)
		os.Remove(gzip_file)
		return false
	}
	if _, err = io.Copy(w, src); err != nil {
		log.Printf("Log: Unable to copy file (rotation) %s: %v", gzip_file, err)
		os.Remove(gzip_file)
		return false
	}
	w.Close()
	return true
}