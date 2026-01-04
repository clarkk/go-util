package serv

import "net/http"

type Adapter func(http.HandlerFunc) http.HandlerFunc

func Adapt(h http.HandlerFunc, wrappers... Adapter) http.HandlerFunc {
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}