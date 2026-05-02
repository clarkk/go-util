package fss

import (
	"io/fs"
	"testing"
	"reflect"
	"path/filepath"
)

var data = []input{
	input {5, "/base/", MIN_DIGITS, "/base"},
	input {45, "/base", MIN_DIGITS, "/base"},
	input {100, "/base", MIN_DIGITS, "/base/100"},
	input {150, "/test", MIN_DIGITS, "/test/100"},
	input {990, "/test", MIN_DIGITS, "/test/900"},
	input {5874, "/test", MIN_DIGITS, "/test/5000/800"},
	input {72000, "/test", MIN_DIGITS, "/test/70000/2000"},
	input {90210600500, "/test", MIN_DIGITS, "/test/90000000000/200000000/10000000/600000/500"},
	input {90390800100, "/test", MIN_DIGITS, "/test/90000000000/300000000/90000000/800000/100"},
	input {90390800200, "/test", MIN_DIGITS, "/test/90000000000/300000000/90000000/800000/200"},
	input {90390800210, "/test", MIN_DIGITS, "/test/90000000000/300000000/90000000/800000/200"},
}

type input struct {
	file_id 	uint64
	path 		string
	min_digits 	int
	output 		string
}

func Test_get(t *testing.T){
	for i := range data {
		got 	:= New(data[i].file_id, data[i].path, data[i].min_digits).Get()
		want 	:= data[i].output
		if got != want {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}

func Test_create_exists_write_fetch_clear(t *testing.T){
	base := t.TempDir()
	
	for i := range data {
		p := New(data[i].file_id, base, data[i].min_digits)
		
		//	Create
		_, err := p.Create()
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		
		//	Exists
		exists, err := p.Exists()
		if err != nil {
			t.Errorf("exists failed: %v", err)
		}
		if !exists {
			t.Errorf("exists failed")
		}
		
		//	Write
		err = p.Write("data1.txt", []byte(""), 0644)
		if err != nil {
			t.Errorf("write failed: %v", err)
		}
		err = p.Write("data2.txt", []byte(""), 0644)
		if err != nil {
			t.Errorf("write failed: %v", err)
		}
		
		//	Fetch
		files, err := p.Fetch()
		if err != nil {
			t.Errorf("fetch failed: %v", err)
		}
		want := []string{
			p.file_prefix()+"data1.txt",
			p.file_prefix()+"data2.txt",
		}
		if !reflect.DeepEqual(files, want) {
			t.Errorf("fetch mismatch %d:\ngot: %v\nwant: %v", data[i].file_id, files, want)
		}
		
		//	Clear
		err = p.Clear(false)
		if err != nil {
			t.Errorf("clear failed: %v", err)
		}
		files, err = p.Fetch()
		if err != nil {
			t.Errorf("fetch failed: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("clear failed: files not cleared")
		}
	}
}

func Test_purge(t *testing.T){
	base := t.TempDir()
	
	p1 := New(90390800200, base, MIN_DIGITS)
	_, err := p1.Create()
	if err != nil {
		t.Errorf("create failed: %v", err)
	}
	
	err = p1.Write("data.txt", []byte(""), 0644)
	if err != nil {
		t.Errorf("write failed: %v", err)
	}
	
	p2 := New(90390800210, base, MIN_DIGITS)
	_, err = p2.Create()
	if err != nil {
		t.Errorf("create failed: %v", err)
	}
	
	p3 := New(90390000000, base, MIN_DIGITS)
	_, err = p3.Create()
	if err != nil {
		t.Errorf("create failed: %v", err)
	}
	
	err = p2.Write("data.txt", []byte(""), 0644)
	if err != nil {
		t.Errorf("write failed: %v", err)
	}
	
	files, err := get_files(base)
	if err != nil {
		t.Errorf("files: %v", err)
	}
	
	want := []string{
		"90000000000",
		"90000000000/300000000",
		"90000000000/300000000/90000000",
		"90000000000/300000000/90000000/800000",
		"90000000000/300000000/90000000/800000/200",
		"90000000000/300000000/90000000/800000/200/90390800200_data.txt",
		"90000000000/300000000/90000000/800000/200/90390800210_data.txt",
	}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("purge mismatch:\ngot: %v\nwant: %v", files, want)
	}
	
	/*err = p1.Clear(true)
	if err != nil {
		t.Errorf("clear failed: %v", err)
	}
	
	want = []string{
		"90000000000",
		"90000000000/300000000",
		"90000000000/300000000/90000000",
		"90000000000/300000000/90000000/800000",
		"90000000000/300000000/90000000/800000/200",
	}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("purge mismatch:\ngot: %v\nwant: %v", files, want)
	}
	
	err = p2.Clear(true)
	if err != nil {
		t.Errorf("clear failed: %v", err)
	}
	
	want = []string{
		"90000000000",
		"90000000000/300000000",
		"90000000000/300000000/90000000",
	}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("purge mismatch:\ngot: %v\nwant: %v", files, want)
	}*/
}

func get_files(root string) ([]string, error){
	var files []string
	root = filepath.Clean(root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel != "." {
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}