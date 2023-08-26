# go-util/serv
Lightweight HTTP server
- With regex patterns in routes
- Bind HTTP methods to routes

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
  
  //  Accept only GET methods
  h.Route(serv.GET, "/get", func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    io.WriteString(w, "This only accepts GET methods")
  })
  
  //  Accept all methods: GET, POST, DELETE etc.
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

## Accept all HTTP methods (GET, POST, DELETE, etc.)
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  path, slugs := serv.Get_path_slugs("", r)
  fmt.Println("path:", path)
  fmt.Println("slugs:", slugs)
  
  io.WriteString(w, "Hello world!")
})
```

## Accept only HTTP POST
```
h.Route(serv.POST, "/post", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  io.WriteString(w, "This route only accepts POST methods!")
})
```

## Regex pattern
All routes will automatically be prefixed with a `^` starting anchor and regex precompiled

`/base_path/([^/]+)` is compiled as `^/base_path/([^/]+)`
```
h.Route(serv.ALL, "/base_path/([^/]+)/test/([^/]+)", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  slug1 := serv.Get_pattern_slug(r, 0)
  slug2 := serv.Get_pattern_slug(r, 1)
  
  io.WriteString(w, "regex path slug names: "+slug1+" "+slug2)
})
```

# go-util/sess
Lightweight HTTP sessions
- With read/write locks to prevent concurrent requests to read data when another is writing data
- Handles all sessions internal in Go to optimize I/O performance
- Uses Redis as failover if Go HTTP server is restarted to preserve and recover sessions

### Example
```
//  Connect to Redis
rdb.Connect(REDIS_AUTH)
//  Initiate session pool and maintenance tasks
sess.Init()

h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session with read-lock
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Get session data
  session_data := s.Get()
  fmt.Println("session data:", session_data)
  
  //  Write data to session
  session_data["test"] = "My data"
  s.Write(data)
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Login (start, regenerate session and close)
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session with read-lock
  s := sess.Start(w, r)
  defer s.Close()
  
  /*
    Add login authentication logic here
  */
  
  //  Regenerate session id after authentication
  s.Regenerate(w)
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Logout (start and destroy session)
```
h.Route(serv.ALL, "/", func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session with read-lock
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Destroy session (will automatically close the session)
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
  session_data := s.Get()
  fmt.Println("session data:", session_data)
  
  //  Write data to session
  session_data["test"] = "My data"
  s.Write(data)
  
  //  Close session as soon as possible to release the read-lock
  s.Close()
  
  io.WriteString(w, "Hello world!")
})
```

## Get session from `r *http.Request` context
```
s = sess.Session(r)
session_data = s.Get()
```