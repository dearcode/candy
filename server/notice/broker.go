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
	log.Debugf("broker message:%v", m)
	g, err := b.gate.client(m.addr)
	if err != nil {
		log.Errorf("connect to %s error:%s", m.addr, err.Error())
		return errors.Trace(err)
	}
	req := &meta.GateNoticeRequest{ChannelID: m.id, Msg: &m.Message}
	_, err = g.Notice(context.Background(), req)

	return errors.Trace(err)
}

func (b *broker) Subscribe(id int64, addr string) {
	b.RLock()
	c, cok := b.channels[id]
	b.RUnlock()

	if !cok {
		c = channel{id: id, gates: make(map[string]gateInfo)}
		b.Lock()
		b.channels[id] = c
		b.Unlock()
	}

	c.RLock()
	_, gok := c.gates[addr]
	c.RUnlock()
	if gok {
		return
	}

	c.Lock()
	c.gates[addr] = gateInfo{addr: addr}
	c.Unlock()
}

func (b *broker) UnSubscribe(id int64, addr string) {
	b.Lock()
	if c, cok := b.channels[id]; cok {
		if _, gok := c.gates[addr]; gok {
			delete(c.gates, addr)
		}
		if len(c.gates) == 0 {
			delete(b.channels, id)
		}
	}
	b.Unlock()
}

func (b *broker) PushOne(msg meta.Message, id int64) {
	log.Debugf("broker msg:%v id:%v", msg, id)
	b.RLock()
	c, cok := b.channels[id]
	b.RUnlock()
	if !cok {
		log.Errorf("user not subscribe, c:%v cok:%v", c, cok)
		return
	}

	c.RLock()
	for _, g := range c.gates {
		b.pusher <- message{id: id, addr: g.addr, Message: msg}
		log.Debugf("pusher msg, id:%v addr:%v msg:%v", id, g.addr, msg)
	}
	c.RUnlock()
}

func (b *broker) Push(msg meta.Message, ids ...int64) {
	log.Debugf("broker msg:%v ids:%v", msg, ids)
	for _, id := range ids {
		b.PushOne(msg, id)
	}
}
