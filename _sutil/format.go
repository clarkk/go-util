package sutil

import (
	"strings"
	"strconv"
)

//	Convert CSV string to []string: "1,2,3" => [1 2 3]
func Fields_csv(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ','
	})
}

//	Convert []int to CSV: [1 2 3] => "1,2,3"
func Int_csv(a []int) string {
	len := len(a)
	if len == 0 {
		return ""
	}
	b := make([]string, len)
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, ",")
}

//	Extract map keys to []string
func Map_keys(m map[string]any) []string {
	a := make([]string, len(m))
	i := 0
	for k := range m {
		a[i] = k
		i++
	}
	return a
}