# go-util
HTTP server with regex patterns as routes

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