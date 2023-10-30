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
	
	RE_SLUG 	= `([^/]+)`
	RE_FILE 	= `([^/]+\.[^/]+)$`
	
	CTX_HTTP 	ctx_key = ""
)

type (
	routes 			[]*route
	
	HTTP struct {
		host 		string
		port 		int
		
		sld 		string
		tld 		string
		
		test 		bool
		
		routes 		routes
	}
	
	route struct {
		method 		string
		pattern 	string
		regex 		*regexp.Regexp
		timeout 	int
		handler		http.HandlerFunc
	}
	
	ctx_key 		string
)

func NewHTTP(host string, port int) *HTTP {
	cutil.Out("Starting server")
	
	sld, tld := parse_directory_host()
	return &HTTP{
		host:		host,
		port:		port,
		sld:		sld,
		tld:		tld,
		routes:		routes{},
	}
}

func Recover(w http.ResponseWriter){
	if r := recover(); r != nil {
		http.Error(w, "Unexpected error", http.StatusBadRequest)
		log.Println(r, "\n"+string(debug.Stack()))
	}
}

func Get_pattern_slug(r *http.Request, index int) string {
	fields := r.Context().Value(CTX_HTTP).([]string)
	return fields[index]
}

func (h *HTTP) Test(){
	cutil.Out("HTTP server in test-mode")
	h.test = true
}

//	Apply route pattern
func (h *HTTP) Route(method string, pattern string, timeout int, handler http.HandlerFunc){
	h.routes = append(h.routes, &route{
		method,
		strip_trailing_slash(pattern),
		nil,
		timeout_min(timeout),
		handler,
	})
}

//	Apply regex route pattern
func (h *HTTP) Route_regex(method string, pattern string, timeout int, handler http.HandlerFunc){
	h.routes = append(h.routes, &route{
		method,
		strip_trailing_slash(pattern),
		regexp.MustCompile("^"+pattern),
		timeout_min(timeout),
		handler,
	})
}

//	Start server
func (h *HTTP) Run(){
	cutil.Out(fmt.Sprintf("Routes defined: %d", len(h.routes)))
	for _, route := range h.routes {
		cutil.Out(route.pattern)
	}
	cutil.Out(fmt.Sprintf("Listening on: %s:%d, SLD: %s, TLD: %s (pid: %d, GOMAXPROCS: %d) running as '%s'",
		h.host,
		h.port,
		h.sld,
		h.tld,
		os.Getpid(),
		runtime.GOMAXPROCS(0),
		cutil.Get_user().Username,
	))
	
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
		
		log.Fatalf("HTTP server: %s", err)
	}
}

//	Route pattern handler
func (h *HTTP) serve(w http.ResponseWriter, r *http.Request){
	var (
		match_route 	*route
		allow 			[]string
	)
	
	ctx := r.Context()
	
	//	Strip trailing slashes from URL
	path := strip_trailing_slash(r.URL.Path)
	
	for _, route := range h.routes {
		if route.regex != nil {
			//	Regex path
			matches := route.regex.FindStringSubmatch(path)
			len 	:= len(matches)
			if len > 0 {
				if route.method != ALL && r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				
				//	Slug group capture
				if len > 1 {
					ctx = context.WithValue(ctx, CTX_HTTP, matches[1:])
				}
				
				match_route = route
				break
			}
		}else{
			//	Path
			if strings.HasPrefix(path, route.pattern) {
				if route.method != ALL && r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				
				match_route = route
				break
			}
		}
	}
	
	if match_route != nil {
		has_timeout := match_route.timeout > 0
		
		//	Apply timeout context
		if has_timeout {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(time.Duration(match_route.timeout) * time.Second))
			
			go func(){
				//	Serve HTTP request to client
				match_route.handler(w, r.WithContext(ctx))
				cancel()
			}()
			
			//	Wait until the context is done/canceled/timeout
			select {
			case <-ctx.Done():
				//	Return HTTP 408 Timeout if request reached timeout
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
				}
			}
			return
		}
		
		//	Serve HTTP request to client
		match_route.handler(w, r.WithContext(ctx))
		return
	}
	
	//	Return HTTP 404/405 if no route was matched
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}else{
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

//	If HTTP server fails to start check which PID is occupying the port
func (h *HTTP) used_port_pid() (string, string){
	c := cutil.Command{}
	c.Run(fmt.Sprintf("netstat -nlp | grep :%d", h.port))
	if !c.Empty(){
		fields := c.Output_fields()[0]
		field := fields[len(fields)-1]
		pid, name, _ := strings.Cut(field, "/")
		
		c = cutil.Command{}
		c.Run(fmt.Sprintf("ps -p %s -o comm=", pid))
		if !c.Empty(){
			name = c.Output_lines()[0]
		}
		
		return pid, name
	}
	return "", ""
}

//	Parse the parent directory name of the go project directory
func parse_directory_host() (string, string) {
	ex, err := os.Executable()
	if err != nil {
		panic("HTTP parse host: "+err.Error())
	}
	path := strings.Split(filepath.Dir(ex), "/")
	path = path[:len(path)-1]
	slug := path[len(path)-1]
	host := strings.Split(slug, ".")
	parts := len(host)
	if parts <= 2 {
		return "", host[0]
	}
	sld := host[:parts-2]
	tld := host[parts-2:]
	return strings.Join(sld[:], "."), strings.Join(tld[:], ".")
}

func timeout_min(timeout int) int {
	if timeout < 0 {
		return 0
	}
	return timeout
}

func strip_trailing_slash(url string) string {
	return strings.TrimRight(url, "/")
}