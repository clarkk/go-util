package fss

import (
	"testing"
	"reflect"
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

func Test_create_exists_write_fetch(t *testing.T){
	path := t.TempDir()
	
	for i := range data {
		p := New(data[i].file_id, path, data[i].min_digits)
		
		_, err := p.Create()
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		
		exists, err := p.Exists()
		if err != nil {
			t.Errorf("exists failed: %v", err)
		}
		
		if !exists {
			t.Errorf("exists failed")
		}
		
		err = p.Write("data1.txt", []byte(""), 0644)
		if err != nil {
			t.Errorf("write failed: %v", err)
		}
		
		err = p.Write("data2.txt", []byte(""), 0644)
		if err != nil {
			t.Errorf("write failed: %v", err)
		}
		
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
	}
}