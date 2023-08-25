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
	COOKIE_NAME 	= "session_token"
	SESSION_HASH 	= "GOREDIS_SESS:%s"
	EXPIRES 		= 60 * 20
	
	ctx_session 	ctx_key = ""
)

type (
	session struct {
		sid 		string
		lock 		sync.Mutex
		closed 		bool
		expires 	int
		data 		session_data
	}
	
	session_data 	map[string]string
	
	ctx_key 		string
)

//	Start session and lock for other concurrent requests to read data from the same session
func Start(p *pool, w http.ResponseWriter, r *http.Request) *session {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	ctx := r.Context()
	
	var (
		sid 	string
		s 		*session
	)
	
	cookie, err := r.Cookie(COOKIE_NAME)
	if err != nil {
		//	Create session cookie and start new session
		sid 	= set_cookie(w)
		s 		= new(p, sid)
	}else{
		sid 	= cookie.Value
		s 		= fetch_session(ctx, p, sid)
		if s == nil {
			//	Create session cookie and start new session
			sid 	= set_cookie(w)
			s 		= new(p, sid)
		}else{
			//	Continue session
			s.reset()
		}
	}
	
	ctx = context.WithValue(ctx, ctx_session, s)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return s
}

//	Get session from request context
func Session(r *http.Request) *session {
	s, ok := r.Context().Value(ctx_session).(*session)
	if !ok {
		return nil
	}
	return s
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
	
	for k, v := range data {
		s.data[k] = v
	}
}

//	Close session for further writes and release read lock
func (s *session) Close(){
	if !s.closed {
		s.closed = true;
		s.lock.Unlock()
		
		//	Update remote session on Redis
		go update_remote_session(context.Background(), s)
	}
}

func fetch_session(ctx context.Context, p *pool, sid string) *session {
	//	Get local session
	local, ok := p.Get(sid)
	if ok {
		if time_unix() > local.expires {
			return nil
		}
		local.lock.Lock()
		return local
	}
	
	//	Get remote session from Redis
	remote, _ := rdb.Get(ctx, fmt.Sprintf(SESSION_HASH, sid))
	if remote != "" {
		//	Copy and use remote session
		s := new(p, sid)
		err := json.Unmarshal([]byte(remote), &s.data)
		if err != nil {
			panic(err)
		}
		return s
	}
	
	return nil
}

func new(p *pool, sid string) *session {
	s := &session{
		sid:		sid,
		expires:	expires(),
		data:		session_data{},
	}
	s.lock.Lock()
	p.Set(sid, s)
	return s
}

func update_remote_session(ctx context.Context, s *session){
	json_bytes, err_json := json.Marshal(s.data)
	if err_json != nil {
		panic(err_json)
	}
	
	if err := rdb.Set(ctx, fmt.Sprintf(SESSION_HASH, s.sid), json_bytes, EXPIRES); err != nil {
		panic(err)
	}
}

func (s *session) reset(){
	s.lock.Lock()
	s.closed 	= false
	s.expires 	= expires()
}

func set_cookie(w http.ResponseWriter) string {
	sid := uuid.NewString()
	//fmt.Println("set cookie sid:", sid)
	serv.Set_cookie(w, COOKIE_NAME, sid, 0)
	return sid
}

func expires() int {
	return time_unix() + EXPIRES
}

func time_unix() int {
	return int(time.Now().Unix())
}