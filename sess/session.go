package sess

import (
	"fmt"
	"time"
	"sync"
	"context"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/clarkk/go-util/serv"
	"github.com/clarkk/go-util/rdb"
)

const (
	DEFAULT_COOKIE_NAME 	= "session_token"
	DEFAULT_REMOTE_HASH 	= "GOREDIS_SESS:%s"
	DEFAULT_EXPIRES 		= 60 * 20
	DEFAULT_PURGE_INTERVAL 	= 5 * time.Minute
	
	CTX_SESSION 			ctx_key = ""
)

var (
	p 				*pool
	once 			sync.Once
	cfg 			*config
)

type (
	Option 				func(*config)
	
	config struct {
		context 		bool
		cookie_name 	string
		remote_hash 	string
		expires 		int
		purge_interval 	time.Duration
	}
	
	session struct {
		sid 			string
		w 				http.ResponseWriter
		lock 			sync.Mutex
		closed 			bool
		expires 		int
		data 			session_data
	}
	
	session_data 		map[string]string
	
	ctx_key 			string
)

//	Enable fetching session from *http.Request context
func Use_context() Option {
	return func(o *config){
		o.context = true
	}
}

func Use_expires(secs int) Option {
	return func(o *config){
		o.expires = secs
	}
}

func Init(opts ...Option){
	once.Do(func(){
		cfg = &config{
			cookie_name:	DEFAULT_COOKIE_NAME,
			remote_hash:	DEFAULT_REMOTE_HASH,
			expires:		DEFAULT_EXPIRES,
			purge_interval:	DEFAULT_PURGE_INTERVAL,
		}
		
		for _, opt := range opts {
			opt(cfg)
		}
		
		p = &pool{
			sessions: sessions{},
		}
		
		//	Purge inactive sessions from pool
		ticker := time.NewTicker(cfg.purge_interval)
		go func(){
			for range ticker.C {
				go p.purge_expired()
			}
		}()
	})
}

//	Start session and lock for other concurrent requests to read data from the same session
func Start(w http.ResponseWriter, r *http.Request) *session {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	ctx := r.Context()
	
	var (
		sid 	string
		s 		*session
	)
	
	cookie, err := r.Cookie(cfg.cookie_name)
	if err != nil {
		//	Create session cookie and start new session
		sid 	= set_cookie(w)
		s 		= new(sid)
	}else{
		sid 	= cookie.Value
		s 		= fetch_session(ctx, sid)
		if s == nil {
			//	Create session cookie and start new session
			sid 	= set_cookie(w)
			s 		= new(sid)
		}else{
			//	Continue session
			s.reset()
		}
	}
	
	s.w = w
	
	if cfg.context {
		ctx = context.WithValue(ctx, CTX_SESSION, s)
		r2 := r.WithContext(ctx)
		*r = *r2
	}
	
	return s
}

//	Fetch session from request context
func Session(r *http.Request) *session {
	if !cfg.context {
		panic("Session context feature is disabled")
	}
	
	s, ok := r.Context().Value(CTX_SESSION).(*session)
	if !ok {
		return nil
	}
	return s
}

//	Regenerate session id
func (s *session) Regenerate(){
	if s.closed {
		panic("Can not regenerate a closed session")
	}
	
	ctx := context.Background()
	
	//	Delete session
	p.delete(s.sid)
	go delete_remote_session(ctx, s.sid)
	
	//	Regenerate and update session
	s.sid = set_cookie(s.w)
	p.set(s.sid, s)
	go update_remote_session(ctx, s)
}

//	Get session ID
func (s *session) Get_sid() string {
	return s.sid
}

//	Get session data
func (s *session) Get() session_data {
	return s.data
}

//	Write session data
func (s *session) Write(data session_data){
	if s.closed {
		panic("Can not write to closed session")
	}
	
	s.data = data
}

//	Close session for further writes and release read lock
func (s *session) Close(){
	if !s.closed {
		s.closed 	= true;
		s.w 		= nil
		s.lock.Unlock()
		go update_remote_session(context.Background(), s)
	}
}

//	Destroy and delete session
func (s *session) Destroy(){
	if s.closed {
		panic("Can not destroy closed session")
	}
	
	s.closed 	= true;
	serv.Delete_cookie(s.w, cfg.cookie_name)
	p.delete(s.sid)
	s.w 		= nil
	s.data 		= nil
	s.lock.Unlock()
	go delete_remote_session(context.Background(), s.sid)
}

func fetch_session(ctx context.Context, sid string) *session {
	//	Get local session
	local, ok := p.get(sid)
	if ok {
		if time_unix() > local.expires {
			return nil
		}
		local.lock.Lock()
		return local
	}
	
	//	Get remote session from Redis
	remote, _ := rdb.Get(ctx, sid_hash(sid))
	if remote != "" {
		//	Copy and use remote session
		s := new(sid)
		err := json.Unmarshal([]byte(remote), &s.data)
		if err != nil {
			panic(err)
		}
		return s
	}
	
	return nil
}

func new(sid string) *session {
	s := &session{
		sid:		sid,
		expires:	expires(),
		data:		session_data{},
	}
	s.lock.Lock()
	p.set(sid, s)
	return s
}

func update_remote_session(ctx context.Context, s *session){
	json_bytes, err_json := json.Marshal(s.data)
	if err_json != nil {
		panic(err_json)
	}
	
	if err := rdb.Set(ctx, fmt.Sprintf(cfg.remote_hash, s.sid), json_bytes, cfg.expires); err != nil {
		panic(err)
	}
}

func delete_remote_session(ctx context.Context, sid string){
	if err := rdb.Delete(ctx, sid_hash(sid)); err != nil {
		panic(err)
	}
}

func (s *session) reset(){
	s.closed 	= false
	s.expires 	= expires()
}

func set_cookie(w http.ResponseWriter) string {
	sid := uuid.NewString()
	serv.Set_cookie(w, cfg.cookie_name, sid, 0)
	return sid
}

func sid_hash(sid string) string {
	return fmt.Sprintf(cfg.remote_hash, sid)
}

func expires() int {
	return time_unix() + cfg.expires
}

func time_unix() int {
	return int(time.Now().Unix())
}