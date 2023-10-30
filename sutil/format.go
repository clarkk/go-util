package sutil

import (
	"strings"
)

func Fields_csv(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ','
	})
}