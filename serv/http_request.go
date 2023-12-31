package serv

import (
	"io"
	"fmt"
	"strings"
	"regexp"
	"net/http"
	//"github.com/go-errors/errors"
)

var regex_get_query = regexp.MustCompile(`^[\pL_][\pL_0-9]+$`)

func Get_client_IP(r *http.Request) string{
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

func Get_path_slugs(r *http.Request, base string) (string, []string){
	path := strings.TrimRight(r.URL.Path, "/")
	if !strings.HasPrefix(path, base) {
		panic("Base path is not a prefix of path")
	}
	return path, strings.Split(strings.TrimLeft(path[len(base):], "/"), "/")
}

//	Check if POST body exceeds limit
func Post_limit(w http.ResponseWriter, r *http.Request, limit_kb int64) ([]byte, error){
	r.Body = http.MaxBytesReader(w, r.Body, limit_kb)
	body_bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body_bytes, nil
}

func Valid_query_param(key string) error {
	if !regex_get_query.MatchString(key) {
		return fmt.Errorf("Invalid query parameter: %s", key)
	}
	return nil
}

func Valid_query_param_single(value []string) (string, bool) {
	switch len(value) {
	//	Single value (empty): OK
	case 0:
		return "", true
	//	Single value: OK
	case 1:
		return value[0], true
	//	Multiple values: ERROR
	default:
		return "", false
	}
}