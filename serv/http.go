package serv

import (
	"log"
	"fmt"
	"os"
	"strings"
	"runtime"
	"runtime/debug"
	"path/filepath"
	"net/http"
	"github.com/clarkk/go-util/cutil"
)

type HTTP struct {
	host 		string
	port 		int
	
	tld 		string
	test 		bool
	
	routes 		map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewHTTP(host string, port int) *HTTP {
	return &HTTP{
		host:	host,
		port:	port,
		tld:	get_tld(),
		routes:	map[string]func(w http.ResponseWriter, r *http.Request){},
	}
}

func Panic_recover(w http.ResponseWriter){
	if r := recover(); r != nil {
		http.Error(w, "Unexpected error", http.StatusBadRequest)
		log.Fatal(r, string(debug.Stack()))
	}
}

func (h *HTTP) Test(){
	cutil.Out("HTTP server in test-mode")
	h.test = true
}

func (h *HTTP) Route(base_url string, handler func(http.ResponseWriter, *http.Request)){
	h.routes[base_url] = handler
}

func (h *HTTP) Run(){
	cutil.Out(fmt.Sprintf("Listening on: %s:%d, %s (pid: %d, GOMAXPROCS: %d) running as '%s'", h.host, h.port, h.tld, os.Getpid(), runtime.GOMAXPROCS(0), cutil.Get_user().Username))
	
	mux := http.NewServeMux()
	for base_url, handler := range h.routes {
		mux.HandleFunc(base_url, handler)
	}
	
	server := &http.Server{
		Addr:		fmt.Sprintf("%s:%d", h.host, h.port),
		Handler:	mux,
	}
	
	err := server.ListenAndServe()
	if err != nil {
		pid, name := h.used_port_pid()
		if pid != "" {
			log.Fatalf("Port %d is already in use by PID %s %s", h.port, pid, name)
		}
		
		log.Fatal("HTTP server: "+err.Error())
	}
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