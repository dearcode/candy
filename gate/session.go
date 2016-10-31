package gate

import (
	"sync"

	"github.com/dearcode/candy/util/log"
)

// sid:Subscribe id.
type session struct {
	user  int64 // 用户ID
	sid   int64
	conns []*connection // 来自不同设备的所有连接
	sync.RWMutex
}

func newSession(id int64, c *connection) *session {
	return &session{user: id, conns: []*connection{c}}
}

func (s *session) addConnection(conn *connection) {
	log.Debugf("%d addr:%s, dev:%s", s.user, conn.getAddr(), conn.getDevice())
	s.Lock()
	s.conns = append(s.conns)
	s.Unlock()
}

// delConnection 遍历session的conns，删除当前连接
func (s *session) delConnection(conn *connection) bool {
	log.Debugf("%d addr:%s, dev:%s", s.user, conn.getAddr(), conn.getDevice())
	empty := false
	s.Lock()
	for i := 0; i < len(s.conns); {
		if s.conns[i].getAddr() == conn.getAddr() {
			copy(s.conns[i:], s.conns[i+1:])
			s.conns = s.conns[:len(s.conns)-1]
			continue
		}
		i++
	}
	if len(s.conns) == 0 {
		empty = true
	}
	s.Unlock()

	return empty
}

//walkConnection 复制遍历
func (s *session) walkConnection(call func(*connection) bool) {
	s.RLock()
	conns := append([]*connection{}, s.conns...)
	s.RUnlock()

	for _, c := range conns {
		if call(c) {
			break
		}
	}
}

func (s *session) setSubscribeID(sid int64) {
	s.sid = sid
}

func (s *session) getSubscribeID() int64 {
	return s.sid
}
