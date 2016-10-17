package gate

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type session struct {
	user  int64         // 用户ID
	conns []*connection // 来自不同设备的所有连接
	sync.Mutex
}

// manager 管理所有session.
type manager struct {
	stop     bool
	notice   *util.Notice
	conns    map[string]*connection // 所有的connection
	sessions map[int64]*session     // 在线用户
	pushChan chan *meta.PushRequest
	sync.RWMutex
}

func newManager() *manager {
	return &manager{sessions: make(map[int64]*session), conns: make(map[string]*connection), pushChan: make(chan *meta.PushRequest, 1000)}
}

func newSession(id int64, c *connection) *session {
	return &session{user: id, conns: []*connection{c}}
}

func (m *manager) addConnection(id int64, device string, c *connection) {
	log.Debugf("addConnection user:%d, conn:%v", id, c)
	c.setDevice(device)

	m.Lock()
	if s, ok := m.sessions[id]; !ok {
		m.sessions[id] = newSession(id, c)
	} else {
		s.addConnection(c)
	}
	m.Unlock()

	//订阅消息
	if err := m.notice.Subscribe(id, device); err != nil {
		log.Errorf("%d dev:%s Subscribe error:%s", id, device, errors.ErrorStack(err))
	}
}

func (m *manager) delConnection(id int64, c *connection) {
	log.Debugf("delConnection user:%d, conn:%v", id, c)

	m.Lock()
	if s, ok := m.sessions[id]; ok {
		s.delConnection(c)
	}
	m.Unlock()

	c.close()

	//消息订阅
	if err := m.notice.Subscribe(id, c.getDevice()); err != nil {
		log.Errorf("UnSubscribe error:%s", errors.ErrorStack(err))
	}
}

func (s *session) addConnection(conn *connection) {
	log.Debugf("%d addConnection:%v", s.user, conn)
	s.Lock()
	s.conns = append(s.conns)
	s.Unlock()
}

// delConnection 遍历session的conns，删除当前连接
func (s *session) delConnection(conn *connection) bool {
	log.Debugf("%d addConnection:%v", s.user, conn)
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
	log.Debugf("%d walkConnection", s.user)
	s.Lock()
	conns := append([]*connection{}, s.conns...)
	s.Unlock()

	for _, c := range conns {
		if call(c) {
			break
		}
	}
}

func (m *manager) getContextAddr(ctx context.Context) (string, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", errors.Trace(ErrInvalidContext)
	}

	addrs, ok := md["remote"]
	if !ok {
		return "", errors.Trace(ErrInvalidContext)
	}

	if len(addrs) != 1 {
		return "", errors.Trace(ErrInvalidContext)
	}

	return addrs[0], nil
}

func (m *manager) getConnection(ctx context.Context) (c *connection, ok bool, err error) {
	var addr string

	if addr, err = m.getContextAddr(ctx); err != nil {
		return
	}

	m.Lock()
	if c, ok = m.conns[addr]; !ok {
		c = newConnection(addr)
		m.conns[addr] = c
	}
	m.Unlock()
	return
}

func (m *manager) getUserSession(id int64) *session {
	m.RLock()
	s, ok := m.sessions[id]
	m.RUnlock()
	if !ok {
		return nil
	}

	return s
}

func (m *manager) getSession(ctx context.Context) (*session, *connection, error) {
	c, ok, err := m.getConnection(ctx)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	if !ok || c.getUser() == 0 {
		return nil, c, errors.Trace(ErrInvalidContext)
	}

    s := m.getUserSession(c.getUser())
	if s == nil {
		return nil, c, errors.Trace(ErrInvalidContext)
	}

	return s, c, nil
}

func (m *manager) start(noticeAddr string) error {
	n, err := util.NewNotice(noticeAddr, m.getPushChan())
	if err != nil {
		return errors.Trace(err)
	}
	m.notice = n

	go m.run()

	return nil
}

func (m *manager) run() {
	for !m.stop {
		req := <-m.pushChan
		for _, id := range req.ID {
			s := m.getUserSession(id.User)
			if s == nil {
				continue
			}
			s.walkConnection(func(c *connection) bool {
				c.send(req.Msg)
				return false
			})
		}
	}
}

func (m *manager) getPushChan() chan<- *meta.PushRequest {
	return m.pushChan
}
