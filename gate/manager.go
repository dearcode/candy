package gate

import (
	"strconv"
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

// manager 管理所有session, connection.
type manager struct {
	stop     bool
	host     string
	notifer  *util.NotiferClient
	conns    map[int64]*connection // 根据token索引
	sessions map[int64]*session    // 根据用户id索引
	pushChan chan meta.PushRequest
	sync.RWMutex
}

func newManager(n *util.NotiferClient, host string) *manager {
	m := &manager{
		notifer:  n,
		host:     host,
		sessions: make(map[int64]*session),
		conns:    make(map[int64]*connection),
		pushChan: make(chan meta.PushRequest, 1000),
	}

	return m
}

func (m *manager) online(id int64, device string, token int64) {
	log.Debugf("user:%d, token:%d, device:%s", id, token, device)

	c := newConnection(id, token, device)

	m.Lock()
	s, ok := m.sessions[id]
	if !ok {
		s = newSession(id, c)
		m.sessions[id] = s
	}

	s.addConnection(c)

	m.conns[token] = c

	m.Unlock()

	//订阅消息
	if err := m.notifer.Subscribe(id, device, token, m.host); err != nil {
		log.Errorf("%d dev:%s Subscribe error:%s", id, device, errors.ErrorStack(err))
		return
	}
}

func (m *manager) offline(c *connection) {
	log.Debugf("user:%d, token:%d, dev:%s", c.getUser(), c.getToken(), c.getDevice())

	m.Lock()
	s, ok := m.sessions[c.getUser()]
	if ok {
		s.delConnection(c)
		delete(m.sessions, c.getUser())
	}

	delete(m.conns, c.getToken())

	m.Unlock()

	//消息订阅
	if err := m.notifer.UnSubscribe(c.getUser(), c.getDevice(), m.host, c.getToken()); err != nil {
		log.Errorf("UnSubscribe error:%s", errors.ErrorStack(err))
	}
}

//getConnByToken 根据int64的token获取对应的connection.
func (m *manager) getConnByToken(token int64) *connection {
	m.Lock()
	c, ok := m.conns[token]
	m.Unlock()

	if ok {
		return c
	}
	return nil
}

func (m *manager) getConnection(ctx context.Context) *connection {
	t, err := util.ContextGet(ctx, "token")
	if err != nil {
		log.Errorf("get token error:%s, ctx:%+v", errors.ErrorStack(err), ctx)
		return nil
	}

	token, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		log.Errorf("ParseInt token:%s error:%s", t, err.Error())
		return nil
	}
	return m.getConnByToken(token)
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

//getSessionByToken 据int64的token获取对应的session.
func (m *manager) getSessionByToken(token int64) *session {
	c := m.getConnByToken(token)
	if c == nil {
		return nil
	}

	return m.getUserSession(c.getUser())
}

func (m *manager) getSession(ctx context.Context) *session {
	c := m.getConnection(ctx)
	if c == nil {
		return nil
	}

	return m.getUserSession(c.getUser())
}

func (m *manager) pushMessage(req *meta.PushRequest) {
	for _, id := range req.ID {
		s := m.getUserSession(id.User)
		if s == nil {
			continue
		}
		s.walkConnection(func(c *connection) bool {
			if err := c.send(&req.Msg); err != nil {
				log.Infof("send Msg:%v, to:%d addr:%s, last:%v, err:%s", req.Msg.Msg.ID, id.User, c.getToken(), c.last, errors.ErrorStack(err))
				m.offline(c)
			}
			return false
		})
	}
}
