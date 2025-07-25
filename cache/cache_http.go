package cache

import (
	"sync"
	"time"
	"bytes"
	"net/http"
)

type (
	Cache_http struct {
		responses	map[string]cache_response
		lock		sync.RWMutex
	}
	
	cache_response struct {
		body		[]byte
		header		http.Header
		code		int
		expires		int64
	}
	
	cache_response_writer struct {
		http.ResponseWriter
		body		*bytes.Buffer
		statusCode	int
	}
)

func New_http(purge_interval int) *Cache_http {
	c := &Cache_http{
		responses: map[string]cache_response{},
	}
	//	Purge expired responses from cache with time interval
	ticker := time.NewTicker(time.Duration(purge_interval) * time.Second)
	go func(){
		for range ticker.C {
			go c.purge_expired()
		}
	}()
	return c
}

func (c *Cache_http) Handler(handler http.HandlerFunc, expires int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		key := r.URL.RequestURI()
		if c.cached_response(w, key) {
			return
		}
		
		c.lock.Lock()
		defer c.lock.Unlock()
		crw := new_cache_response_writer(w)
		handler.ServeHTTP(crw, r)
		
		//	Cache response
		if crw.statusCode == http.StatusOK {
			c.responses[key] = cache_response{
				body:		crw.body.Bytes(),
				header:		copy_headers(w),
				code:		crw.statusCode,
				expires:	time_expires(expires),
			}
		}
	})
}

func (c *Cache_http) cached_response(w http.ResponseWriter, key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	entry, found := c.responses[key]
	if !found {
		return false
	}
	//	Check if response has expired
	if entry.expires != 0 && time_unix() > entry.expires {
		return false
	}
	//	Write cached headers
	for k, values := range entry.header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(entry.code)
	w.Write(entry.body)
	return true
}

func (c *Cache_http) purge_expired(){
	c.lock.Lock()
	defer c.lock.Unlock()
	time_unix := time_unix()
	for key, i := range c.responses {
		if i.expires != 0 && time_unix > i.expires {
			delete(c.responses, key)
		}
	}
}

func copy_headers(w http.ResponseWriter) http.Header {
	headers	:= w.Header()
	copied	:= make(http.Header, len(headers))
	for k, values := range headers {
		values_copied := make([]string, len(values))
		copy(values_copied, values)
		copied[k] = values_copied
	}
	return copied
}

func new_cache_response_writer(w http.ResponseWriter) *cache_response_writer {
	return &cache_response_writer{
		ResponseWriter:	w,
		body:			new(bytes.Buffer),
		statusCode:		http.StatusOK,
	}
}

func (w *cache_response_writer) WriteHeader(code int){
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *cache_response_writer) Write(b []byte) (int, error){
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}