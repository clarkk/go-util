# go-util/serv
Lightweight HTTP server
- With regex patterns in routes
- Bind HTTP methods to routes

```
package main

import (
  "fmt"
  "io"
  "net/http"
  "github.com/clarkk/go-util/serv"
)

func main(){
  h := serv.NewHTTP("127.0.0.1", 3000)
  
  //  Accepts only GET methods
  h.Route(serv.GET, "/get", func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    io.WriteString(w, "This only accepts GET methods")
  })
  
  //  Accepts only POST methods
  h.Route(serv.POST, "/post", func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    io.WriteString(w, "This only accepts POST methods")
  })
  
  //  Use regex in route pattern
  h.Route(serv.ALL, "/base_path/([^/]+)/test/([^/]+)", func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    slug1 := serv.Get_pattern_slug(r, 0)
    slug2 := serv.Get_pattern_slug(r, 1)
    
    io.WriteString(w, "regex path slug names: "+slug1+" "+slug2)
  })
  
  //  Accepts all methods: GET, POST, DELETE etc.
  h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    path, slugs := serv.Get_path_slugs("", r)
    fmt.Println("path:", path)
    fmt.Println("slugs:", slugs)
    
    io.WriteString(w, "Hello world!")
  })
  
  h.Run()
}
```

# go-util/sess
Lightweight HTTP sessions
- With read/write locks to prevent concurrent requests to read data when another is writing data
- Handles all sessions internal in Go to optimize I/O performance
- Uses Redis as failover if Go HTTP server is restarted to preserve and recover sessions

```
//  Connect to Redis
rdb.Connect(REDIS_AUTH)
//  Start session pool
sessions := sess.Init()

h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session with read-lock
  s := sess.Start(sessions, w, r)
  defer s.Close()
  
  //  Get session data
  session_data := s.Get()
  fmt.Println("session data:", session_data)
  
  //  Write data to session
  session_data["test"] = "My data"
  s.Write(data)
  
  //  Get session from request context
  s = sess.Session(r)
  session_data = s.Get()
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```