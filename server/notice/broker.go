package notice

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util/log"
)

type gateInfo struct {
	addr string
}

type channel struct {
	id    int64
	gates map[string]gateInfo
	sync.RWMutex
}

type message struct {
	id   int64
	addr string
	meta.Message
}

type broker struct {
	channels map[int64]channel
	sync.RWMutex

	pusher chan message
	gate   *gate
}

const (
	defaultPushChanSize = 1000
)

func newBroker() *broker {
	return &broker{channels: make(map[int64]channel), pusher: make(chan message, defaultPushChanSize), gate: newGate()}
}

// Start start service.
func (b *broker) Start() error {
	go func() {
		for {
			m := <-b.pusher
			log.Debugf("sendMessage msg:%v", m)
			b.sendMessage(m)
		}
	}()
	return nil
}

func (b *broker) sendMessage(m message) error {
	log.Debugf("broker message:%v, to gate:%s", m, m.addr)
	g, err := b.gate.client(m.addr)
	if err != nil {
		log.Errorf("connect to %s error:%s", m.addr, err.Error())
		return errors.Trace(err)
	}
	req := &meta.GateNoticeRequest{ChannelID: m.id, Msg: &m.Message}
	log.Debugf("begin call gate Notice")
	resp, err := g.Notice(context.Background(), req)
	log.Debugf("end call gate Notice err:%v, head:%v", err, resp.Header.Error())
	if err != nil {
		log.Errorf("Notice to gate:%s error:%s", m.addr, err.Error())
		return errors.Trace(err)
	}

	return errors.Trace(err)
}

func (b *broker) Subscribe(id int64, addr string) {
	b.Lock()

	c, ok := b.channels[id]
	if !ok {
		c = channel{id: id, gates: make(map[string]gateInfo)}
		b.channels[id] = c
	}

	if _, ok = c.gates[addr]; !ok {
		c.gates[addr] = gateInfo{addr: addr}
	}

	b.Unlock()

	log.Debugf("Subscribe id:%d, addr:%s", id, addr)
}

func (b *broker) UnSubscribe(id int64, addr string) {
	b.Lock()
	if c, ok := b.channels[id]; ok {
		delete(c.gates, addr)
		delete(b.channels, id)
	}
	b.Unlock()
	log.Debugf("UnSubscribe id:%d, addr:%s", id, addr)
}

func (b *broker) Push(msg meta.Message, ids ...int64) {
	log.Debugf("broker msg:%v ids:%v", msg, ids)
	b.Lock()

	for _, id := range ids {
		if c, ok := b.channels[id]; ok {
			for _, gate := range c.gates {
				b.pusher <- message{id: id, addr: gate.addr, Message: msg}
				log.Debugf("pusher msg, id:%v addr:%v msg:%v", id, gate.addr, msg)
			}
		}
	}

	b.Unlock()
}
