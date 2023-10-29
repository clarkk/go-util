package serv

import "net/http"

type Adapter func(h http.Handler) http.Handler

func Adapt(h http.HandlerFunc, wrappers ...Adapter) http.HandlerFunc {
	for _, wrapper := range wrappers {
		h = wrapper(h).ServeHTTP
	}
	return h
}