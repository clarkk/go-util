package serv

import (
	"strings"
	"net/http"
)

func Get_path_slugs(base string, r *http.Request) (string, []string){
	path := strings.TrimRight(r.URL.Path, "/")
	if !strings.HasPrefix(path, base) {
		panic("Base path is not a prefix of path")
	}
	return path, strings.Split(strings.TrimLeft(path[len(base):], "/"), "/")
}