# Install
`go get -u github.com/clarkk/go-util`

All packages are extremely simple and lightweight by design

- [go-util/cache](#go-utilcache) Cache with TTL
- [go-util/hash_pass](#go-utilhash_pass) Secure password hashing for storing passwords in databases etc.
- [go-util/lang](#go-utillang) Multi-lingual translations with both strings and errors
- [go-util/serv](#go-utilserv) HTTP server
- [go-util/sess](#go-utilsess) HTTP sessions

# go-util/cache
Lightweight cache with expires (TTL) and syncronized with `sync.RWMutex` that ensures only one can write to the cache at the time.

### Example
```
package main

import (
  "fmt"
  "github.com/clarkk/go-util/cache"
)

const (
  //  Set the cache expires to 1 hour
  expires = 60 * 60
  
  //  Set the interval when to purge expired values in the cache to 1 minute
  purge_interval = 60
)

//  Declares a cache with string values
var cache_string *cache.Cache[string]

func main(){
  //  Create cache
  cache_string = cache.New[string](purge_interval)
  
  cache_key := "key-to-cached-value"
  s, found := cache_string.Get(cache_key)
  //  Check if the value is cached
  if !found {
    //  Cache the value with 1 hour expiration
    s = "Value to cache"
    cache_string.Set(cache_key, s, expires)
  }
  
  fmt.Println("Cached value:", s)
}
```

# go-util/hash_pass
Secure password hashing for storing passwords in databases etc.
- Hashing with Argon2id algorithm
- Adding randomly generated salt value to hash generation
- Configuration settings are included in the hashing string including the random salt value
- Hashing string will be a fixed length (depending on the hasing settings)

### Example
```
package main

import (
  "log"
  "fmt"
  "time"
  "github.com/clarkk/go-util/hash_pass"
)

func main(){
  password := "the-password"
  
  //  Create new hashing string
  hash, err := hash_pass.Create(password)
  if err != nil {
    log.Fatal(err)
  }
  
  fmt.Println("hash:", hash)
  
  //  Compare if a password is equal to hashing string
  valid, err := hash_pass.Compare(password, hash)
  if err != nil {
    log.Fatal(err)
  }
  
  if valid {
    fmt.Println("The password is correct")
  } else {
    //  Adding sleep if password is incorrect
    time.Sleep(2 * time.Second)
    
    fmt.Println("The password is incorrect")
  }
}
```

## Hashing configuration options
```
const (
  time uint32       = 1
  memory uint32     = 1024 * 64
  salt_bytes uint32 = 32
  hash_bytes uint32 = 128
)
```

# go-util/lang
Handle multiple languages with both errors and strings.

This package uses `go-util/cache` to cache translations to improve performance.

## Interface to fetch translations from external source like database etc.
```
type Adapter interface {
  Fetch(lang, table, key string) (string, error)
}
```

## Database structure
```
CREATE TABLE `lang` (
  `id` smallint(5) UNSIGNED NOT NULL AUTO_INCREMENT,
  `sid` varchar(50) NOT NULL,
  `da` varchar(400) NOT NULL,
  `en` varchar(400) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `sid` (`sid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_general_ci;

CREATE TABLE `lang_error` (
  `id` smallint(5) UNSIGNED NOT NULL AUTO_INCREMENT,
  `sid` varchar(50) NOT NULL,
  `da` varchar(100) NOT NULL,
  `en` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `sid` (`sid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_general_ci;

INSERT INTO `lang` (`sid`, `da`, `en`) VALUES
  ('HELLO_WORLD', 'Hej verden', 'Hello world'),
  ('WELCOME', 'Hej %name% og velkommen', 'Hi %name% and welcome');

INSERT INTO `lang_error` (`sid`, `da`, `en`) VALUES
  ('FIELD_INT_RANGE', 'Skal være mellem %min% til %max%', 'Must be between %min% and %max%');
```

### Example
```
package main

import (
  "fmt"
  "errors"
  "database/sql"
  "github.com/clarkk/go-util/lang"
  "github.com/clarkk/go-util/serv/req"
)

//  This struct is satisfied by the interface 
type lang_fetch struct {}

func (l lang_fetch) Fetch(lang, table, key string) (string, error){
  var s string
  if err := fetch_from_db("SELECT "+lang+" FROM ."+table+" WHERE sid=?", key, &s); err != nil {
    s = key
    if !errors.Is(err, sql.ErrNoRows) {
      //  Return fatal errors
      return s, err
    }
  }
  return s, nil
}

func main(){
  supported_langs := []string{
    "da",
    "en",
  }
  
  //  Set cache expires to 24 hours
  expires := 60 * 60 * 24
  
  //  Initiate
  lang.Init(lang_fetch{}, expires, supported_langs)
}

func route_handler(w http.ResponseWriter, r *http.Request){
  //  Optional to set a specific language
  language := "en"
  
  //  Optional to get 'Accept-Language' header if provided by the client in request
  accept_lang := req.Accept_lang(r)
  
  //  Create language instance
  l := lang.New(language, accept_lang)
  
  fmt.Println(l.String("HELLO_WORLD", nil))
  
  fmt.Println(l.String("WELCOME", lang.Rep{
    "name": "Stephen"
  }))
  
  err := l.Error("FIELD_INT_RANGE", lang.Rep{
    "min": 1,
    "max": 100,
  })
  fmt.Println("Error:", err)
}
```

# go-util/serv
Lightweight HTTP server
- Shutdown gracefully on SIGINT (CTRL+C or "kill -INT $pid")
- Handles subdomains
- With regex pattern in routes (placeholders)
- Bind HTTP methods to routes
- Set individual timeout on each route (with `context.WithTimeout()` on request handler)
- Supports customizable adapters/middleware

Route pattern types on subdomain
- `Route_exact()`: Only matches the exact URL
- `Route()`: Matches all URL's with the pattern prefix
- `Route_blind()`: Matches all URL's with the pattern prefix and returns HTTP 404

Route pattern placeholders (regex)
- `:slug` = `[^/]+`
- `:file` = `[a-z\d\-_]+\.[a-z]{1,4}`

All incoming HTTP requests will have trailing slashes trimmed before matching with route pattern: `/foo/bar/` => `/foo/bar`

## Use nginx as reverse proxy
TLS is too slow in Go because the user no longer has access to define TLS ciphers etc.
Nginx is much more performant in handling the TLS encryption between the client and the Go server, where you have freedom to choose TLS ciphers and other TLS related configuration to improve performance.
```
server {
  listen  80;
  server_name  subdomain.domain.com;
  
  return  301 https://$server_name$request_uri;
}

server {
  listen  443 ssl;
  http2  on;
  server_name  subdomain.domain.com;
  
  ssl_certificate  /var/ssl/domain.com/fullchain.pem;
  ssl_certificate_key  /var/ssl/domain.com/private.key;
  
  error_log  /var/log/nginx/error_subdomain.domain.com.log warn;
  
  location / {
    proxy_pass  http://127.0.0.1:8000/;
    proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
    proxy_set_header  Host  $host;
    proxy_request_buffering  off;
  }
}
```

### Example
```
package main

import (
  "fmt"
  "io"
  "net/http"
  "github.com/clarkk/go-util/serv"
  "github.com/clarkk/go-util/serv/req"
)

//  Middleware executed before the main HTTP handler
func adapt_method1() serv.Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
      //  This recover method must be called in the first chained handler
      //  in every route to prevent the server from crashing in case of a panic
      defer serv.Recover(w)
      
      //  This handler is executed in a chain before the main HTTP handler
      
      fmt.Println("Method1 executed")
      
      //  Execute next handler in the chain
      h.ServeHTTP(w, r)
    })
  }
}

//  Middleware executed before the main HTTP handler
func adapt_method2() serv.Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
      //  This handler is executed in a chain before the main HTTP handler
      
      fmt.Println("Method2 executed")
      
      //  Execute next handler in the chain
      h.ServeHTTP(w, r)
    })
  }
}

func main(){
  //  Initiate server
  h := serv.NewHTTP("domain.com", "127.0.0.1", 8000)
  
  //  Subdomain: subdomain.domain.com
  h.Subhost("subdomain.").
    
    //  Accepts only GET methods (with 60 seconds timeout)
    Route(serv.GET, "/get", 60, serv.Adapt(
      func(w http.ResponseWriter, r *http.Request){
        fmt.Println("Main handler executed")
        
        io.WriteString(w, "This only accepts GET methods")
      },
      
      //  Apply a chain of adapters/middleware before the main HTTP handler
      adapt_method1(),
      adapt_method2(),
    )).
    
    //  Accepts all methods: GET, POST, DELETE etc. (with 60 seconds timeout)
    Route(serv.ALL, "/", 60, serv.Adapt(
      func(w http.ResponseWriter, r *http.Request){
        fmt.Println("Main handler executed")
        
        path, slugs := req.Get_path_slugs(r, "")
        fmt.Println("path:", path)
        fmt.Println("slugs:", slugs)
        
        io.WriteString(w, "Hello world!")
      },
      
      //  Apply a chain of adapters/middleware before the main HTTP handler
      adapt_method1(),
      adapt_method2(),
    ))
  
  //  Start server
  h.Run()
}
```

## Accepts all HTTP methods (GET, POST, DELETE, etc.) with 60 second timeout
```
Route(serv.ALL, "/", 60, func(w http.ResponseWriter, r *http.Request){
  //  This recover method must be called in the first chained handler
  //  in every route to prevent the server from crashing in case of a panic
  defer serv.Recover(w)
  
  path, slugs := serv.Get_path_slugs(r, "")
  fmt.Println("path:", path)
  fmt.Println("slugs:", slugs)
  
  io.WriteString(w, "Hello world!")
})
```

## Accepts only HTTP POST with 120 second timeout
```
Route(serv.POST, "/post", 120, func(w http.ResponseWriter, r *http.Request){
  //  This recover method must be called in the first chained handler
  //  in every route to prevent the server from crashing in case of a panic
  defer serv.Recover(w)
  
  io.WriteString(w, "This route only accepts POST methods!")
})
```

## Accepts only HTTP GET and only matches the exact pattern with 60 second timeout
```
Route_exact(serv.GET, "/exact-url/file.ext", 60, func(w http.ResponseWriter, r *http.Request){
  //  This recover method must be called in the first chained handler
  //  in every route to prevent the server from crashing in case of a panic
  defer serv.Recover(w)
  
  io.WriteString(w, "This route is exact!")
})
```

## Blind route pattern (HTTP 404)
```
Route_blind(serv.GET, "/http404")
```

## Regex route pattern (placeholders) with 60 second timeout

### Placeholder for slugs
```
Route(serv.ALL, "/base_path/:slug/test/:slug", 60, func(w http.ResponseWriter, r *http.Request){
  //  This recover method must be called in the first chained handler
  //  in every route to prevent the server from crashing in case of a panic
  defer serv.Recover(w)
  
  slug1 := serv.Get_slug(r, 0)
  slug2 := serv.Get_slug(r, 1)
  
  io.WriteString(w, "slug1: "+slug1+", slug2: "+slug2)
})
```

### Placeholder for file with exact route
Matches all files in the given directory
```
Route_exact(serv.ALL, "/base_path/:file", 60, func(w http.ResponseWriter, r *http.Request){
  //  This recover method must be called in the first chained handler
  //  in every route to prevent the server from crashing in case of a panic
  defer serv.Recover(w)
  
  io.WriteString(w, "This is a file!")
})
```

## Custom adapters/middleware
```
//  Verify user authentication
func adapt_auth() serv.Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
      //  This recover method must be called in the first chained handler
      //  in every route to prevent the server from crashing in case of a panic
      defer serv.Recover(w)
      
      //  Verify Auth Basic
      if !serv.Auth_basic(AUTH_USER, AUTH_PASS) {
        io.WriteString(w, "No access!")
        return
      }
      
      h.ServeHTTP(w, r)
    })
  }
}

Route(serv.GET, "/get", 60, serv.Adapt(
    //  Executed last
    func(w http.ResponseWriter, r *http.Request){
      io.WriteString(w, "Hello world!")
    },
    
    //  Apply a chain of adapters/middleware before the main HTTP handler
    adapt_auth(),             //  Executed first
    adapt_something(),        //  Executed second
    adapt_something_else(),   //  Executed third
  ))
```

# go-util/sess
Lightweight HTTP sessions
- With read/write lock (`sync.RWMutex`) to prevent concurrent requests to read/write to the same session data
- Handles all sessions internal in Go to improve I/O performance
- Uses Redis as failover if Go HTTP server is restarted/crashed to preserve and recover sessions

### Example
```
import (
  "github.com/clarkk/go-util/rdb"
  "github.com/clarkk/go-util/sess"
)

//  Connect to Redis
rdb.Connect(REDIS_HOST, REDIS_HOST, REDIS_AUTH)

//  Initiate session pool and maintenance tasks
sess_expires = 60 * 20
sess_cookie_name = "session_token"
sess_redis_prefix = "GOREDIS_SESS"
sess_purge_expired = 60
sess.Init(sess_expires, sess_cookie_name, sess_redis_prefix, sess_purge_expired)

h.Route(serv.ALL, "/", 60, func(w http.ResponseWriter, r *http.Request){
  defer serv.Recover(w)
  
  //  Start session (with read-lock)
  s := sess.Start(w, r)
  defer s.Close()
  
  //  Get session data
  data := s.Data()
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
  data := s.Data()
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
s := sess.Request(r)
data := s.Data()
```