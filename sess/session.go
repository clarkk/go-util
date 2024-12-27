package sess

import (
	"sync"
	"time"
	"context"
	"net/http"
	"github.com/google/uuid"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-util/rdb"
	"github.com/clarkk/go-util/serv"
)

const (
	ctx_sess ctx_key 	= ""
)

var (
	once 					sync.Once
	p 						*pool
	session_expires 		int
	session_cookie_name 	string
	session_remote_prefix	string
)

type (
	Session struct {
		sid 		string
		w 			http.ResponseWriter
		lock 		sync.Mutex
		closed 		bool
		expires 	int64
		data 		session_data
	}
	
	sessions 		map[string]*Session
	
	session_data 	map[string]any
	
	ctx_key 		string
)

func Init(expires int, cookie_name, remote_prefix string, purge_interval int){
	once.Do(func(){
		session_expires 		= expires
		session_cookie_name		= cookie_name
		session_remote_prefix	= remote_prefix
		
		p = &pool{
			sessions: sessions{},
		}
		
		//	Purge inactive sessions from pool
		ticker := time.NewTicker(time.Duration(purge_interval) * time.Second)
		go func(){
			for range ticker.C {
				go p.purge_expired()
			}
		}()
	})
}

//	Start session and lock for other concurrent requests to read data from the same session
func Start(w http.ResponseWriter, r *http.Request) *Session {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	ctx := r.Context()
	
	var (
		sid 	string
		s 		*Session
	)
	
	cookie, err := r.Cookie(session_cookie_name)
	if err != nil {
		//	Create session cookie and start new session
		sid 	= set_cookie(w)
		s 		= create_session(sid)
	} else {
		sid 	= cookie.Value
		s 		= fetch_session(ctx, sid)
		if s == nil {
			//	Create session cookie and start new session
			sid 	= set_cookie(w)
			s 		= create_session(sid)
		} else {
			//	Continue session
			s.reset()
		}
	}
	
	s.w = w
	
	ctx = context.WithValue(ctx, ctx_sess, s)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return s
}

//	Fetch session from request context
func Request(r *http.Request) *Session {
	s, ok := r.Context().Value(ctx_sess).(*Session)
	if !ok {
		return nil
	}
	return s
}

//	Regenerate session id
func (s *Session) Regenerate(){
	if s.Closed() {
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
func (s *Session) Sid() string {
	return s.sid
}

//	Check if session data is empty
func (s *Session) Empty() bool {
	return len(s.data) == 0
}

//	Check if session is closed
func (s *Session) Closed() bool {
	return s.closed
}

//	Get session data
func (s *Session) Data() session_data {
	if s.csrf_token() == "" {
		return s.data
	}
	data := s.data
	delete(data, csrf_token)
	return data
}

//	Write session data
func (s *Session) Write(data map[string]any){
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	if _, ok := data[csrf_token]; ok {
		panic("Can not use reserved CSRF key in session")
	}
	
	s.data = data
}

//	Close session for further writes and release read lock
func (s *Session) Close(){
	if s.close() {
		go update_remote_session(context.Background(), s)
	}
}

//	Destroy and delete session
func (s *Session) Destroy(){
	if s.Closed() {
		panic("Can not destroy closed session")
	}
	s.closed 	= true;
	serv.Delete_cookie(s.w, session_cookie_name)
	if s.csrf_token() != "" {
		serv.Delete_cookie(s.w, csrf_token)
	}
	p.delete(s.sid)
	s.w 		= nil
	s.data 		= nil
	s.lock.Unlock()
	go delete_remote_session(context.Background(), s.sid)
}

func (s *Session) csrf_token() string {
	if token, ok := s.data[csrf_token]; ok {
		return token.(string)
	}
	return ""
}

func (s *Session) close() bool {
	if s.Closed() {
		return false
	}
	s.closed 	= true;
	s.w 		= nil
	s.lock.Unlock()
	return true
}

func (s *Session) reset(){
	s.closed 	= false
	s.expires 	= expires()
}

func create_session(sid string) *Session {
	s := &Session{
		sid:		sid,
		expires:	expires(),
		data:		session_data{},
	}
	s.lock.Lock()
	p.set(sid, s)
	return s
}

func fetch_session(ctx context.Context, sid string) *Session {
	//	Get local session
	s, expired := p.get(sid);
	if expired {
		p.delete(sid)
		return nil
	}
	if s != nil {
		return s
	}
	
	//	Get remote session from Redis
	if remote, _ := rdb.Get(ctx, sid_hash(sid)); remote != "" {
		//	Copy and use remote session
		s := create_session(sid)
		if err := json.Unmarshal([]byte(remote), &s.data); err != nil {
			panic("Session remote fetch JSON decode: "+err.Error())
		}
		return s
	}
	
	return nil
}

func update_remote_session(ctx context.Context, s *Session){
	b, err := json.Marshal(s.data)
	if err != nil {
		panic("Session remote update JSON encode: "+err.Error())
	}
	if err := rdb.Set(ctx, sid_hash(s.sid), b, session_expires); err != nil {
		panic("Session remote update: "+err.Error())
	}
}

func delete_remote_session(ctx context.Context, sid string){
	if err := rdb.Del(ctx, sid_hash(sid)); err != nil {
		panic("Session remote delete: "+err.Error())
	}
}

func set_cookie(w http.ResponseWriter) string {
	sid := uuid_string()
	serv.Set_cookie(w, session_cookie_name, sid, 0)
	return sid
}

func sid_hash(sid string) string {
	return session_remote_prefix+":"+sid
}

func expires() int64 {
	return time_unix() + int64(session_expires)
}

func time_unix() int64 {
	return time.Now().Unix()
}

func uuid_string() string {
	return uuid.NewString()
}