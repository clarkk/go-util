package serv

import (
	"os"
	"os/signal"
	"log"
	"fmt"
	"time"
	"sync"
	"regexp"
	"strings"
	"context"
	"runtime"
	"net/http"
	"github.com/go-errors/errors"
	"github.com/clarkk/go-util/cmd"
)

const ctx_slug ctx_key = ""

var re_sld = regexp.MustCompile(`^[a-z]+(?:[a-z\-]+[a-z]+)?\.$`)

type (
	HTTP struct {
		tld 		string
		tld_len 	int
		listen_ip 	string
		listen_port int
		test		bool
		subhosts 	subhosts
	}
	
	subhosts 		map[string]*subhost
	
	ctx_key 		string
)

func NewHTTP(tld, listen_ip string, listen_port int) *HTTP {
	cmd.Out("Initiating HTTP server: "+tld)
	
	return &HTTP{
		tld:			tld,
		tld_len:		len(tld),
		listen_ip:		listen_ip,
		listen_port:	listen_port,
		subhosts:		subhosts{},
	}
}

//	Recover from panic inside route handler
func Recover(w http.ResponseWriter){
	if err := recover(); err != nil {
		if !w.(*Writer).Sent_headers() {
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
		}
		log.Println(errors.Wrap(err, 2).ErrorStack())
	}
}

func Get_slug(r *http.Request, index int) string {
	value := r.Context().Value(ctx_slug)
	if value == nil {
		return ""
	}
	fields := value.([]string)
	if len(fields) < index + 1 {
		return ""
	}
	return fields[index]
}

func (h *HTTP) Test(){
	cmd.Out("HTTP server in test-mode")
	h.test = true
}

//	Apply subhost with underlying routes
func (h *HTTP) Subhost(sld string) *subhost {
	if !re_sld.MatchString(sld) {
		if sld[len(sld)-1:] != "." {
			log.Fatalf("Subhost must end with '.': %s -> %s.", sld, sld)
		}
		log.Fatalf("Subhost must only contain a-z and '-': %s", sld)
	}
	if _, ok := h.subhosts[sld]; ok {
		log.Fatalf("Subhost already exists: %s", sld)
	}
	h.subhosts[sld] = &subhost{
		map_routes:		map_routes{},
		map_exact:		map_exact{},
		routes:			routes{},
	}
	return h.subhosts[sld]
}

//	Start server
func (h *HTTP) Run(){
	h.output_init()
	
	srv := &http.Server{
		Addr:				fmt.Sprintf("%s:%d", h.listen_ip, h.listen_port),
		Handler:			http.HandlerFunc(h.serve),
		//ReadTimeout:		5 * time.Second,
		ReadHeaderTimeout:	100 * time.Millisecond,
		//WriteTimeout:		10 * time.Second,
		IdleTimeout:		30 * time.Second,
	}
	
	wait := &sync.WaitGroup{}
	wait.Add(1)
	
	go func(){
		defer wait.Done()
		
		//	Always returns http.ErrServerClosed on SIGINT
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//	Server stopped unexpectedly
			if pid, name := h.used_port_pid(); pid != "" {
				log.Fatalf("Port %d is already in use by PID %s %s",
					h.listen_port,
					pid,
					name,
				)
			}
			log.Fatalf("HTTP server: %s", err)
		}
		
		cmd.Out("HTTP server stopped serving")
	}()
	
	//	Listening for SIGINT to shutdown gracefully (stop accepting new connections/requests): CTRL+C or "kill -INT $pid"
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	
	cmd.Out("HTTP server received SIGINT to shutdown gracefully")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown: %s", err)
	}
	
	wait.Wait()
	
	cmd.Out("HTTP server shutdown completed successfully")
}

//	Subhost and route pattern handler
func (h *HTTP) serve(w http.ResponseWriter, r *http.Request){
	w = &Writer{ResponseWriter: w}
	fmt.Println("host:", r.Host, "tld:")
	if !strings.HasSuffix(r.Host, h.tld) || r.Host == h.tld {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		log.Printf("Unsupported host (TLD %s): %s", h.tld, r.Host)
		return
	}
	
	sld 	:= r.Host[:len(r.Host)-h.tld_len]
	fmt.Println("sld:", sld)
	s, ok 	:= h.subhosts[sld]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		log.Printf("Unsupported subhost (SLD %s): %s", sld, r.Host)
		return
	}
	
	ctx 	:= r.Context()
	path 	:= strip_trailing_slash(r.URL.Path)
	fmt.Println("path:", path)
	var match_route *route_handler
	for _, route := range s.routes {
		if route.regex != nil {
			//	Regex pattern
			matches := route.regex.FindStringSubmatch(path)
			len 	:= len(matches)
			if len > 0 {
				if route.depth != 0 && !match_path_depth(path, matches[0]) {
					continue
				}
				
				handler, ok := match_method(route, w, r)
				if !ok {
					return
				}
				
				//	Slug group capture
				if len > 1 {
					ctx = context.WithValue(ctx, ctx_slug, matches[1:])
				}
				
				match_route = handler
				break
			}
		} else {
			//	Path
			if match_path(path, route) {
				handler, ok := match_method(route, w, r)
				if !ok {
					return
				}
				match_route = handler
				break
			}
		}
	}
	
	//	Return HTTP 404 if no route was matched or route is blind
	if match_route == nil || match_route.blind {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	
	//	Apply timeout context
	if match_route.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(match_route.timeout) * time.Second)
		defer cancel()
		
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
	} else {
		//	Serve HTTP request to client
		match_route.handler(w, r.WithContext(ctx))
	}
}

func (h *HTTP) output_init(){
	cmd.Outf("Subhosts: %d\n", len(h.subhosts))
	for sld, s := range h.subhosts {
		cmd.Outf("%s%s (Routes: %d)\n", sld, h.tld, len(s.routes))
		for _, route := range s.routes {
			var pattern string
			if route.regex != nil {
				pattern = fmt.Sprintf("%s", route.regex)
			} else {
				pattern = route.pattern
				if route.exact {
					pattern = "="+pattern
				}
			}
			
			cmd.Out(pattern)
			for method, handler := range route.methods {
				if handler.blind {
					cmd.Outf("\t%s HTTP 404\n", method)
				} else {
					cmd.Outf("\t%s %d\n", method, handler.timeout)
				}
			}
		}
	}
	
	usr, _ := cmd.Get_user()
	
	cmd.Outf("Listening on: %s:%d, TLD: %s (PID: %d, GOMAXPROCS: %d) running as '%s'\n",
		h.listen_ip,
		h.listen_port,
		h.tld,
		os.Getpid(),
		runtime.GOMAXPROCS(0),
		usr.Username,
	)
}

//	If HTTP server fails to start check which PID is occupying the port
func (h *HTTP) used_port_pid() (string, string){
	c := cmd.Command{}
	c.Run(fmt.Sprintf("netstat -nlp | grep :%d", h.listen_port))
	if !c.Empty(){
		fields 	:= c.Output_fields()[0]
		field 	:= fields[len(fields)-1]
		pid, name, _ := strings.Cut(field, "/")
		
		c = cmd.Command{}
		c.Run(fmt.Sprintf("ps -p %s -o comm=", pid))
		if !c.Empty(){
			name = c.Output_lines()[0]
		}
		return pid, name
	}
	return "", ""
}

func match_method(route *route, w http.ResponseWriter, r *http.Request) (*route_handler, bool){
	if handler, ok := route.methods[string(ALL)]; ok {
		return handler, true
	}
	
	handler, ok := route.methods[r.Method]
	if !ok {
		//	Return HTTP 405 if no route was matched
		methods := make([]string, len(route.methods))
		i := 0
		for k := range route.methods {
			methods[i] = k
			i++
		}
		w.Header().Set("Allow", strings.Join(methods, ", "))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil, false
	}
	return handler, true
}

func match_path(path string, route *route) bool {
	if route.exact {
		return path == route.pattern
	}
	if !strings.HasPrefix(path, route.pattern) {
		return false
	}
	if route.depth != 0 {
		return match_path_depth(path, route.pattern)
	}
	return true
}

func match_path_depth(path, pattern string) bool {
	return path == pattern || path[len(pattern)] == '/'
}

func strip_trailing_slash(url string) string {
	if url == "/" {
		return url
	}
	return strings.TrimRight(url, "/")
}