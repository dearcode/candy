package gate

import (
	"time"
)

const (
	stateOffline = iota
	stateOnline
)

type session struct {
	id    int64
	state int
	last  int64
}

func newSession() *session {
	return &session{}
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
