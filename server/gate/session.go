package gate

import (
	"time"

	"github.com/ngaut/log"
)

const (
	stateOffline = iota
	stateOnline
)

type session struct {
	id    int64
	state int
	last  int64
	addr  string
}

func newSession(addr string) *session {
	log.Debugf("new session from:%s", addr)
	return &session{addr: addr}
}

func (s *session) online(id int64) {
	s.state = stateOnline
	s.id = id
	s.last = time.Now().Unix()
}

func (s *session) offline() {
	s.state = stateOffline
}

func (s *session) update() {
	s.last = time.Now().Unix()
}

func (s *session) getAddr() string {
	return s.addr
}

func (s *session) getID() int64 {
	return s.id
}

func (s *session) isOnline() bool {
	return s.state == stateOnline
}
