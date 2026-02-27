package sess

import (
	"log"
	"fmt"
	"sync"
	"time"
	"context"
	"net/http"
	"encoding/json/v2"
	"github.com/google/uuid"
	"github.com/clarkk/go-util/rdb"
	"github.com/clarkk/go-util/serv"
)

const ctx_sess ctx_key = ""

var (
	once 					sync.Once
	p 						*pool
	session_expires 		int
	session_cookie_name 	string
	session_remote_prefix	string
)

type (
	Session struct {
		closed 		bool
		w 			http.ResponseWriter
		r 			*http.Request
		data 		session_data
		sess 		*session
	}
	
	sessions 		map[string]*session
	session struct {
		sid 		string
		lock 		sync.Mutex
		expires 	int64
		data 		session_data
	}
	
	session_data struct {
		Keys		map[string]any	`json:"keys"`
		Csrf_token	string			`json:"csrf_token"`
	}
	
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
				func(){
					defer func(){
						if r := recover(); r != nil {
							log.Printf("purge_expired panic: %v", r)
						}
					}()
					p.purge_expired()
				}()
			}
		}()
	})
}

//	Start session and lock for other concurrent requests to read data from the same session
func Start(w http.ResponseWriter, r *http.Request) (*Session, error){
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	ctx := r.Context()
	
	var (
		sid 	string
		sess 	*session
		err 	error
	)
	
	cookie, err := r.Cookie(session_cookie_name)
	if err != nil {
		//	Create session cookie and start new session
		sid 		= set_cookie(w)
		sess 		= create_session(sid)
	} else {
		sid 		= cookie.Value
		sess, err 	= fetch_session(ctx, sid)
		if err != nil {
			return nil, err
		}
		if sess == nil {
			//	Create session cookie and start new session
			sid 	= set_cookie(w)
			sess 	= create_session(sid)
		} else {
			//	Continue session
			sess.reset()
		}
	}
	
	s := wrap_session(sess)
	
	ctx = context.WithValue(ctx, ctx_sess, s)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	s.w = w
	s.r = r
	
	return s, nil
}

//	Fetch session from request context
func Request(r *http.Request) *Session {
	s, _ := r.Context().Value(ctx_sess).(*Session)
	return s
}

//	Regenerate session id
func (s *Session) Regenerate(){
	if s.Closed() {
		panic("Can not regenerate a closed session")
	}
	
	ctx := context.Background()
	
	//	Delete session
	p.delete(s.sess.sid)
	go delete_remote_session(ctx, s.sess.sid)
	
	//	Regenerate sid and update session
	s.sess.sid = set_cookie(s.w)
	p.set(s.sess.sid, s.sess)
	go update_remote_session(ctx, s.sess)
}

//	Get session ID
func (s *Session) Sid() string {
	if s.Closed() {
		panic("Can not fetch session id on a closed session")
	}
	
	return s.sess.sid
}

//	Check if session data is empty
func (s *Session) Empty() bool {
	return len(s.data.Keys) == 0
}

//	Check if session is closed
func (s *Session) Closed() bool {
	return s.closed
}

//	Get session data
func (s *Session) Data() map[string]any {
	return copy_data(s.data.Keys)
}

//	Write session data
func (s *Session) Write(data map[string]any){
	if s.Closed() {
		panic("Can not write to closed session")
	}
	
	copied := copy_data(data)
	
	s.data.Keys			= copied
	s.sess.data.Keys	= copied
}

//	Re-open session, write and close
func (s *Session) Write_back(data map[string]any) error {
	sess, expired := p.get(s.sess.sid);
	if sess == nil || expired {
		return fmt.Errorf("Session expired")
	}
	
	//	Write
	for k, v := range data {
		s.data.Keys[k]		= v
		sess.data.Keys[k]	= v
	}
	
	//	Close
	sess.lock.Unlock()
	go update_remote_session(context.Background(), sess)
	
	return nil
}

//	Close session for further writes and release read lock
func (s *Session) Close(){
	if s.close() {
		go update_remote_session(context.Background(), s.sess)
	}
}

//	Destroy and delete session
func (s *Session) Destroy(){
	if s.Closed() {
		panic("Can not destroy closed session")
	}
	
	s.closed 			= true;
	s.data.Keys 		= nil
	s.data.Csrf_token	= ""
	
	//	Delete session
	p.delete(s.sess.sid)
	go delete_remote_session(context.Background(), s.sess.sid)
	serv.Delete_cookie(s.w, session_cookie_name)
	if s.csrf_token() != "" {
		serv.Delete_cookie(s.w, csrf_token)
	}
	
	s.sess 		= nil
}

func (s *Session) csrf_token() string {
	return s.data.Csrf_token
}

func (s *Session) close() bool {
	if s.Closed() {
		return false
	}
	s.closed	= true;
	s.data.Keys	= copy_data(s.data.Keys)
	s.sess.lock.Unlock()
	return true
}

func (s *session) reset(){
	s.expires = expires()
}

func create_session(sid string) *session {
	s := &session{
		sid:		sid,
		expires:	expires(),
		data:		session_data{
			Keys: map[string]any{},
		},
	}
	s.lock.Lock()
	p.set(sid, s)
	return s
}

func fetch_session(ctx context.Context, sid string) (*session, error){
	//	Get local session
	s, expired := p.get(sid);
	if expired {
		p.delete(sid)
		return nil, nil
	}
	if s != nil {
		return s, nil
	}
	
	//	Get remote session from Redis
	remote, not_found, err := rdb.Get(ctx, sid_hash(sid))
	if err != nil && !not_found {
		return nil, err
	}
	if remote != "" {
		//	Copy and use remote session
		s := create_session(sid)
		if err := json.Unmarshal([]byte(remote), &s.data); err != nil {
			s.lock.Unlock()
			panic("Session remote fetch JSON decode: "+err.Error())
		}
		return s, nil
	}
	
	return nil, nil
}

func update_remote_session(ctx context.Context, s *session){
	defer func(){
		if r := recover(); r != nil {
			log.Printf("update_remote_session panic: %v", r)
		}
	}()
	
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

func copy_data(data map[string]any) map[string]any {
	copied := make(map[string]any, len(data))
	for k, v := range data {
		copied[k] = v
	}
	return copied
}

func wrap_session(s *session) *Session {
	return &Session{
		data:	s.data,
		sess:	s,
	}
}

func set_cookie(w http.ResponseWriter) string {
	sid := uuid_string()
	serv.Set_cookie_session(w, session_cookie_name, sid)
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