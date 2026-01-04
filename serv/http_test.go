package serv

import (
	"fmt"
	"slices"
	"strings"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	sld 		= "subdomain."
	tld 		= "domain.com"
	base_url	= sld+tld
)

type args struct {
	req		*http.Request
}

func Test_routes(t *testing.T){
	tests := []struct{
		name		string
		args		func(t *testing.T) args
		want_code	int
		want_body	string
	}{
		{
			name: "invalid TLD",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, "unknown.unknown.com/get")}
			},
			want_code: http.StatusNotFound,
			want_body: http.StatusText(http.StatusNotFound),
		},
		{
			name: "invalid SLD",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, "unknown."+tld+"/get")}
			},
			want_code: http.StatusNotFound,
			want_body: http.StatusText(http.StatusNotFound),
		},
		{
			name: "path 1",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/test")}
			},
			want_code: http.StatusOK,
			want_body: "/",
		},
		{
			name: "path 2",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/get")}
			},
			want_code: http.StatusOK,
			want_body: "/get",
		},
		{
			name: "path 3",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/get-more")}
			},
			want_code: http.StatusOK,
			want_body: "/",
		},
		{
			name: "path 4",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/get-more/more")}
			},
			want_code: http.StatusOK,
			want_body: "/",
		},
		{
			name: "path file",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/file/file.json")}
			},
			want_code: http.StatusOK,
			want_body: "/file/file.json",
		},
		{
			name: "invalid path file base dir",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/file/file.xml")}
			},
			want_code: http.StatusOK,
			want_body: "/file",
		},
		{
			name: "path invalid method",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/post")}
			},
			want_code: http.StatusMethodNotAllowed,
			want_body: http.StatusText(http.StatusMethodNotAllowed),
		},
		{
			name: "regex slug 1",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/regex/match")}
			},
			want_code: http.StatusOK,
			want_body: "/regex/:slug",
		},
		{
			name: "regex slug 2",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/regex/match/test")}
			},
			want_code: http.StatusOK,
			want_body: "/regex/:slug",
		},
		{
			name: "regex file 1",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/regex/file.json")}
			},
			want_code: http.StatusOK,
			want_body: "/regex/:file",
		},
		{
			name: "regex invalid method",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/regex-post/match")}
			},
			want_code: http.StatusMethodNotAllowed,
			want_body: http.StatusText(http.StatusMethodNotAllowed),
		},
		{
			name: "blind 1",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/blind/base/test")}
			},
			want_code: http.StatusOK,
			want_body: "/blind/base/test",
		},
		{
			name: "blind 2",
			args: func(t *testing.T) args {
				return args{test_request(t, http.MethodGet, base_url+"/blind/base")}
			},
			want_code: http.StatusNotFound,
			want_body: http.StatusText(http.StatusNotFound),
		},
	}
	
	h := NewHTTP(tld, "", 0)
	
	h.Subhost(sld).
		Route_exact(GET, "/file/file.json", 0, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/file/file.json")
		}).
		Route(GET, "/file", 0, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/file")
		}).
		Route_exact(GET, "/regex/:file", 60, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/regex/:file")
		}).
		Route_exact(GET, "/blind/base/test", 60, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/blind/base/test")
		}).
		Route_blind(ALL, "/blind").
		Route(GET, "/get", 0, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/get")
		}).
		Route(POST, "/post/", 0, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/post/")
		}).
		Route(ALL, "/regex/:slug", 60, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/regex/:slug")
		}).
		Route(POST, "/regex-post/:slug", 60, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/regex-post/:slug")
		}).
		Route(GET, "/", 0, func(w http.ResponseWriter, r *http.Request){
			fmt.Fprint(w, "/")
		})
	
	handler := h.test_handler()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.args(t).req)
			
			code := w.Result().StatusCode
			body := strings.TrimSpace(w.Body.String())
			
			if code != tt.want_code {
				t.Fatalf("HTTP code want: [%d], but got [%d] %s",
					tt.want_code,
					code,
					body,
				)
			}
			
			if body != tt.want_body {
				t.Fatalf("HTTP response want [%s] but got [%s]",
					tt.want_body,
					body,
				)
			}
		})
	}
}

func Test_priority_routing(t *testing.T){
	h := NewHTTP(tld, "", 0)
	
	handler := func(w http.ResponseWriter, r *http.Request){}
	
	h.Subhost(sld).
		Priority_routing().
			Route_exact(GET, "/file/file.json", 0, handler).
			Route(GET, "/file", 0, handler).
			Route_exact(GET, "/regex/:file", 0, handler).
			Route_exact(GET, "/blind/base/test", 0, handler).
			Route(GET, "/regex/slug", 0, handler).
			Route_blind(GET, "/blind").
			Route(GET, "/get", 0, handler).
			Route(GET, "/regex/slug/path", 0, handler).
			Route_exact(GET, "/get/test/path", 0, handler).
			Route(GET, "/get/test", 0, handler).
			Route(GET, "/post/", 0, handler).
			Route(GET, "/regex/:slug", 0, handler).
			Route(GET, "/regex-post/:slug", 0, handler).
			Route(GET, "/", 0, handler)
	
	s := h.subhosts[sld]
	s.sort_priority()
	
	var got []string
	for _, route := range s.routes {
		got = append(got, route.string())
	}
	
	want := []string{
		"=/blind/base/test",
		"=/get/test/path",
		"=/file/file.json",
		" /get/test",
		"=/regex/"+re_file_pattern,
		" /regex-post/"+re_slug_pattern,
		" /regex/slug/path",
		" /regex/slug",
		" /regex/"+re_slug_pattern,
		" /blind",
		" /get",
		" /file",
		" /post",
		" /",
	}
	
	if !slices.Equal(want, got) {
		t.Fatalf("Failed to sort priority routes\nwant:\n%s\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

func (h *HTTP) test_handler() http.HandlerFunc {
	return http.HandlerFunc(h.serve)
}

func test_request(t *testing.T, method, url string) *http.Request {
	req, err := http.NewRequest(method, "//"+url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	return req
}