package serv

import (
	"log"
	"fmt"
	"os"
	"time"
	"strings"
	"regexp"
	"context"
	"runtime"
	"runtime/debug"
	"path/filepath"
	"net/http"
	"github.com/clarkk/go-util/cutil"
)

const (
	ALL 		= "*"
	GET 		= "GET"
	POST 		= "POST"
	DELETE 		= "DELETE"
	
	ctx_http 	ctx_key = ""
)

type (
	HTTP struct {
		host 		string
		port 		int
		
		tld 		string
		test 		bool
		
		routes 		[]*route
	}
	
	route struct {
		method 		string
		path 		string
		regex 		*regexp.Regexp
		handler		http.HandlerFunc
	}
	
	ctx_key 		string
)

func NewHTTP(host string, port int) *HTTP {
	return &HTTP{
		host:	host,
		port:	port,
		tld:	get_tld(),
		routes:	[]*route{},
	}
}

func Recover(w http.ResponseWriter){
	if r := recover(); r != nil {
		http.Error(w, "Unexpected error", http.StatusBadRequest)
		log.Println(r, "\n"+string(debug.Stack()))
	}
}

func Get_pattern_slug(r *http.Request, index int) string {
	fields := r.Context().Value(ctx_http).([]string)
	return fields[index]
}

func (h *HTTP) Test(){
	cutil.Out("HTTP server in test-mode")
	h.test = true
}

func (h *HTTP) Route(method string, pattern string, handler http.HandlerFunc){
	h.routes = append(h.routes, &route{
		method,
		pattern,
		nil,
		handler,
	})
}

func (h *HTTP) Route_regex(method string, pattern string, handler http.HandlerFunc){
	h.routes = append(h.routes, &route{
		method,
		"",
		regexp.MustCompile("^"+pattern),
		handler,
	})
}

func (h *HTTP) Run(){
	cutil.Out(fmt.Sprintf("Listening on: %s:%d, %s (pid: %d, GOMAXPROCS: %d) running as '%s'", h.host, h.port, h.tld, os.Getpid(), runtime.GOMAXPROCS(0), cutil.Get_user().Username))
	
	srv := &http.Server{
		Addr:				fmt.Sprintf("%s:%d", h.host, h.port),
		Handler:			http.HandlerFunc(h.serve),
		//ReadTimeout:		5 * time.Second,
		ReadHeaderTimeout:	100 * time.Millisecond,
		//WriteTimeout:		10 * time.Second,
		IdleTimeout:		30 * time.Second,
	}
	
	err := srv.ListenAndServe()
	if err != nil {
		pid, name := h.used_port_pid()
		if pid != "" {
			log.Fatalf("Port %d is already in use by PID %s %s", h.port, pid, name)
		}
		
		log.Fatal("HTTP server: "+err.Error())
	}
}

func (h *HTTP) serve(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range h.routes {
		if route.regex != nil {
			//	Regex path
			matches := route.regex.FindStringSubmatch(r.URL.Path)
			if len(matches) > 0 {
				if route.method != ALL && r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				
				ctx := context.WithValue(r.Context(), ctx_http, matches[1:])
				route.handler(w, r.WithContext(ctx))
				return
			}
		}else{
			//	Path
			if strings.HasPrefix(r.URL.Path, route.path) {
				if route.method != ALL && r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				
				route.handler(w, r)
				return
			}
		}
	}
	
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *HTTP) used_port_pid() (string, string){
	c := &cutil.Command{}
	c.Run(fmt.Sprintf("netstat -nlp | grep :%d", h.port))
	if !c.Empty(){
		fields := c.Output_fields()[0]
		field := fields[len(fields)-1]
		pid, name, _ := strings.Cut(field, "/")
		
		c = &cutil.Command{}
		c.Run(fmt.Sprintf("ps -p %s -o comm=", pid))
		if !c.Empty(){
			name = c.Output_lines()[0]
		}
		
		return pid, name
	}
	return "", ""
}

func get_tld() string{
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := strings.Split(filepath.Dir(ex), "/")
	path = path[:len(path)-1]
	slug := path[len(path)-1]
	host := strings.Split(slug, ".")
	host = host[len(host)-2:]
	return strings.Join(host[:], ".")
}