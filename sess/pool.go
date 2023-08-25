package sess

import (
	"sync"
	"time"
)

const (
	purge_interval = 5
)

type pool struct {
	lock 		sync.RWMutex
	sessions 	map[string]*session
}

func Init() *pool {
	p := &pool{
		sessions:	map[string]*session{},
	}
	
	//	Purge inactive sessions from pool
	ticker := time.NewTicker(purge_interval * time.Minute)
	go func(){
		for range ticker.C {
			go p.purge_expired()
		}
	}()
	
	return p
}

func (p *pool) Set(sid string, s *session){
	p.lock.Lock()
	defer p.lock.Unlock()
	
	p.sessions[sid] = s
}

func (p *pool) Get(sid string) (*session, bool){
	p.lock.RLock()
	defer p.lock.RUnlock()
	
	s, ok := p.sessions[sid]
	return s, ok
}

func (p *pool) Delete(sid string){
	p.lock.Lock()
	defer p.lock.Unlock()
	
	delete(p.sessions, sid)
}

func (p *pool) purge_expired(){
	p.lock.Lock()
	defer p.lock.Unlock()
	
	time_unix := time_unix()
	for sid, s := range p.sessions {
		if time_unix > s.expires {
			delete(p.sessions, sid)
		}
	}
}