package gate

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

// manager 管理所有session.
type manager struct {
	stop     bool
	notifer  *util.Notifer
	conns    map[string]*connection // 所有的connection
	sessions map[int64]*session     // 在线用户
	pushChan chan meta.PushRequest
	sync.RWMutex
}

func newManager(host string) (*manager, error) {
	n, err := util.NewNotifer(host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	m := &manager{
		notifer:  n,
		sessions: make(map[int64]*session),
		conns:    make(map[string]*connection),
		pushChan: make(chan meta.PushRequest, 1000),
	}

	go m.run()

	return m, nil
}

func (m *manager) online(id int64, device string, c *connection) {
	log.Debugf("user:%d, conn:%+v, device:%s", id, c, device)
	c.setDevice(device)

	m.Lock()
	s, ok := m.sessions[id]
	if !ok {
		s = newSession(id, c)
		m.sessions[id] = s
	}

	s.addConnection(c)

	m.conns[c.getAddr()] = c

	m.Unlock()

	c.setUser(id)

	//订阅消息
	if err := m.notifer.Subscribe(id, device); err != nil {
		log.Errorf("%d dev:%s Subscribe error:%s", id, device, errors.ErrorStack(err))
	}
}

func (m *manager) offline(id int64, c *connection) {
	log.Debugf("user:%d, conn:%+v", id, c)

	m.Lock()
	if s, ok := m.sessions[id]; ok {
		s.delConnection(c)
	}

	delete(m.conns, c.getAddr())

	m.Unlock()

	c.close()

	//消息订阅
	if err := m.notifer.Subscribe(id, c.getDevice()); err != nil {
		log.Errorf("UnSubscribe error:%s", errors.ErrorStack(err))
	}
}

func (m *manager) getConnection(ctx context.Context) (c *connection, ok bool, err error) {
	var addr string

	if addr, err = util.ContextAddr(ctx); err != nil {
		return
	}
	log.Debugf("conn from:%s, conns:%+v", addr, m.conns)
	m.Lock()
	if c, ok = m.conns[addr]; !ok {
		c = newConnection(addr)
		m.conns[addr] = c
		log.Debugf("conn from:%s, c:%+v", addr, c)
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
		return nil, c, errors.Trace(util.ErrInvalidContext)
	}

	s := m.getUserSession(c.getUser())
	if s == nil {
		return nil, c, errors.Trace(util.ErrInvalidContext)
	}

	return s, c, nil
}

func (m *manager) run() {
	for !m.stop {
		req, err := m.notifer.Recv()
		if err != nil {
			log.Errorf("recv push message error:%s", err)
			continue
		}
		for _, id := range req.ID {
			s := m.getUserSession(id.User)
			if s == nil {
				continue
			}
			s.walkConnection(func(c *connection) bool {
				c.send(&req.Msg)
				return false
			})
		}
	}
}
