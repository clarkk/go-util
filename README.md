# Install
`go get -u github.com/clarkk/go-util`

# go-util/serv
Lightweight HTTP server
- With regex pattern in routes
- Bind HTTP method(s) to routes
- Set individual timeout on each route (with `context.WithTimeout()` on request handler)

### Example
```
package main

import (
  "fmt"
  "log"
  "io"
  "net/http"
  "github.com/clarkk/go-util/serv"
)

func init(){
  log.SetFlags(log.LstdFlags | log.Llongfile | log.Lmicroseconds)
}

func main(){
  h := serv.NewHTTP("127.0.0.1", 3000)
  
  //  Accepts only GET methods (with 60 seconds timeout)
  h.Route(serv.GET, "/get", 60, func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    io.WriteString(w, "This only accepts GET methods")
  })
  
  //  Accepts all methods: GET, POST, DELETE etc. (with 60 seconds timeout)
  h.Route(serv.ALL, "/", 60, func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    path, slugs := serv.Get_path_slugs(r, "")
    fmt.Println("path:", path)
    fmt.Println("slugs:", slugs)
    
    io.WriteString(w, "Hello world!")
  })
  
  h.Run()
}
```

## Accepts all HTTP methods (GET, POST, DELETE, etc.) with 60 second timeout
```
h.Route(serv.ALL, "/", 60, func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  path, slugs := serv.Get_path_slugs(r, "")
  fmt.Println("path:", path)
  fmt.Println("slugs:", slugs)
  
  io.WriteString(w, "Hello world!")
})
```

## Accepts only HTTP POST with 120 second timeout
```
h.Route(serv.POST, "/post", 120, func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  io.WriteString(w, "This route only accepts POST methods!")
})
```

## Regex route pattern with 60 second timeout
All regex routes will automatically be precompiled and prefixed with a `^` starting anchor

`/base_path/([^/]+)` is compiled as `^/base_path/([^/]+)`
```
h.Route_regex(serv.ALL, "/base_path/([^/]+)/test/([^/]+)", 60, func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  slug1 := serv.Get_pattern_slug(r, 0)
  slug2 := serv.Get_pattern_slug(r, 1)
  
  io.WriteString(w, "regex path slug names: "+slug1+" "+slug2)
})
```

# go-util/sess
Lightweight HTTP sessions
- With read/write lock to prevent concurrent requests to read/write to the same session data
- Handles all sessions internal in Go to improve I/O performance
- Uses Redis as failover if Go HTTP server is restarted/crashed to preserve and recover sessions

### Example
```
import (
  "github.com/clarkk/go-util/rdb"
  "github.com/clarkk/go-util/sess"
)

//  Connect to Redis
rdb.Connect(REDIS_AUTH)

//  Initiate session pool and maintenance tasks
sess.Init()

h.Route(serv.ALL, "/", 60, func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session (with read-lock)
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Get session data
  data := s.Get()
  fmt.Println("session data:", data)
  
  //  Write data to session
  data["test"] = "My data"
  s.Write(data)
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Login (start, regenerate session id and close)
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session (with read-lock)
  s := sess.Start(w, r)
  defer s.Close()
  
  /*
    Add login authentication logic here
  */
  
  //  Regenerate session id after authentication
  s.Regenerate()
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Logout (start and destroy session)
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session (with read-lock)
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Destroy session (will close the session)
  s.Destroy()
  
  io.WriteString(w, "Hello world!")
})
```

## Start, write and close
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session with read-lock
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Get session data
  data := s.Get()
  fmt.Println("session data:", data)
  
  //  Write data to session
  data["test"] = "My data"
  s.Write(data)
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Get session from `r *http.Request` context
```
s := sess.Session(r)
data = s.Get()
```