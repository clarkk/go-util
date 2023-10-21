package sess

import (
	"sync"
)

type (
	sessions 			map[string]*session
	
	pool struct {
		lock 			sync.RWMutex
		sessions 		sessions
	}
)

func (p *pool) set(sid string, s *session){
	p.lock.Lock()
	defer p.lock.Unlock()
	p.sessions[sid] = s
}

func (p *pool) get(sid string) (*session, bool){
	p.lock.RLock()
	defer p.lock.RUnlock()
	s, ok := p.sessions[sid]
	return s, ok
}

func (p *pool) delete(sid string){
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