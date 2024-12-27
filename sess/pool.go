package sess

import "sync"

type pool struct {
	lock 			sync.RWMutex
	sessions 		sessions
}

func (p *pool) get(sid string) (*session, bool){
	p.lock.RLock()
	defer p.lock.RUnlock()
	s, ok := p.sessions[sid]
	if !ok {
		return nil, false
	}
	s.lock.Lock()
	//	Check if session is expired
	if time_unix() > s.expires {
		s.lock.Unlock()
		return nil, true
	}
	return s, false
}

func (p *pool) set(sid string, s *session){
	p.lock.Lock()
	defer p.lock.Unlock()
	p.sessions[sid] = s
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