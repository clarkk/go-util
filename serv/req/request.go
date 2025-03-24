package req

import (
	"io"
	"strings"
	"net/http"
)

func Get_client_IP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	ip, _, _ = strings.Cut(ip, ":")
	return ip
}

//	Get all path slugs in URL
func Get_path_slugs(r *http.Request, base string) (string, []string){
	path := strings.TrimRight(r.URL.Path, "/")
	if !strings.HasPrefix(path, base) {
		panic("Base path is not a prefix of path")
	}
	return path, strings.Split(strings.TrimLeft(path[len(base):], "/"), "/")
}

//	Check if POST body exceeds limit
func Post_limit(w http.ResponseWriter, r *http.Request, limit_kb int){
	r.Body = http.MaxBytesReader(w, r.Body, int64(limit_kb * 1024))
}

//	Check if POST body exceeds limit and read body
func Post_limit_read(w http.ResponseWriter, r *http.Request, limit_kb int) ([]byte, error){
	Post_limit(w, r, limit_kb)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//	Get accepted languages
func Accept_lang(r *http.Request) []string {
	s := r.Header.Get("Accept-Language")
	list := []string{}
	unique := map[string]bool{}
	for _, v := range strings.Split(s, ",") {
		lang, _, found := strings.Cut(v, ";")
		if found {
			v = lang
		}
		lang, _, found = strings.Cut(v, "-")
		if found {
			v = lang
		}
		v = strings.TrimSpace(v)
		if v == "*" {
			continue
		}
		if _, found := unique[v]; found {
			continue
		}
		unique[v] = true
		list = append(list, strings.ToLower(v))
	}
	return list
}

//	Check if GET query param is set but empty
func Query_param_empty(s []string) bool {
	if len(s) > 1 || s[0] != "" {
		return false
	}
	return true
}