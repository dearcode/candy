package notice

import (
	"sync"

	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type message struct {
	ids  []*meta.PushID
	addr string
	meta.Message
}

type sender interface {
	notice(string, []*meta.PushID, *meta.Message) error
}

// users 存储用户ID对应的gate地址
type broker struct {
	mbox  chan message
	gate  sender
	users map[int64]string
	sync.RWMutex
}

const (
	defaultPushChanSize = 1000
	// 默认推送线程数量
	defaultSenderNumber = 4
)

func newBroker(gate sender) *broker {
	return &broker{users: make(map[int64]string), mbox: make(chan message, defaultPushChanSize), gate: gate}
}

// Start start service.
func (b *broker) Start() {
	for i := 0; i < defaultSenderNumber; i++ {
		go b.sender(i)
	}
}

// split 按用户所在gate拆分请求
func (b *broker) split(m message) []*message {
	//数量少，不适合用map
	var msgs []*message

	b.RLock()
	for _, id := range m.ids {
		addr, ok := b.users[id.User]
		if !ok {
			// 跳过未订阅的用户
			continue
		}

		msg := (*message)(nil)
		for i, m := range msgs {
			if m.addr == addr {
				msg = msgs[i]
				break
			}
		}

		if msg == nil {
			msg = &message{addr: addr}
			msgs = append(msgs, msg)
		}

		msg.ids = append(msg.ids, id)
	}
	b.RUnlock()

	return msgs
}

func (b *broker) sender(sid int) {
	for {
		m := <-b.mbox

		for _, msg := range b.split(m) {
			log.Debugf("%d begin sendMessage:%v, to gate:%s", sid, m, msg.addr)
			if err := b.gate.notice(msg.addr, msg.ids, &m.Message); err != nil {
				log.Errorf("%d sendMessage error:%s", sid, errors.ErrorStack(err))
				continue
			}
			log.Debugf("%d end sendMessage:%v, to gate:%s", sid, m, msg.addr)
		}
	}
}

func (b *broker) Subscribe(id int64, addr string) {
	b.Lock()
	if _, ok := b.users[id]; !ok {
		b.users[id] = addr
	}
	b.Unlock()
	log.Debugf("Subscribe id:%d, addr:%s", id, addr)
}

func (b *broker) UnSubscribe(id int64) {
	b.Lock()
	delete(b.users, id)
	b.Unlock()

	log.Debugf("UnSubscribe id:%d", id)
}

// Push 过滤掉未订阅用户
func (b *broker) Push(msg meta.Message, ids ...*meta.PushID) {
	log.Debugf("broker msg:%v ids:%v", msg, ids)
	b.mbox <- message{ids: ids, Message: msg}
}
