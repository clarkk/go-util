package serv

import "net/http"

type Adapter func(http.HandlerFunc) http.HandlerFunc

func Adapt(h http.HandlerFunc, wrappers... Adapter) http.HandlerFunc {
	for _, wrapper := range wrappers {
		h = wrapper(h)
	}
	return h
}