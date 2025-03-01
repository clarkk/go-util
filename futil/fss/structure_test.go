package fss

import "testing"

var data = []input{
	input {5, "/base/", 2, "/base"},
	input {45, "/base", 2, "/base"},
	input {100, "/base", 2, "/base/100"},
	input {150, "/test", 2, "/test/100"},
	input {990, "/test", 2, "/test/900"},
	input {5874, "/test", 2, "/test/5000/800"},
	input {72000, "/test", 2, "/test/70000/2000"},
	input {90390800100, "/test", 2, "/test/90000000000/300000000/90000000/800000/100"},
}

type input struct {
	file_id 	int
	path 		string
	min_digits 	int
	output 		string
}

func Test_get(t *testing.T){
	for i := range data {
		got 	:= Get(data[i].file_id, data[i].path, data[i].min_digits)
		want 	:= data[i].output
		if got != want {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}