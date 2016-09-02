package notice

import (
	"sync"

	"github.com/juju/errors"
	"github.com/ngaut/log"

	"github.com/dearcode/candy/server/meta"
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
	for {
		m := <-b.pusher
		b.sendMessage(m)
	}
}

func (b *broker) sendMessage(m message) error {
	g, err := b.gate.client(m.addr)
	if err != nil {
		log.Errorf("connect to %s error:%s", m.addr, err.Error())
		return errors.Trace(err)
	}
	req := &meta.NoticeRequest{ChannelID: m.id, Msg: &m.Message}
	_, err = g.Notice(nil, req)

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

func (b *broker) Unsubscribe(id int64, addr string) {
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

func (b *broker) Push(id int64, msg meta.Message) {
	b.RLock()
	c, cok := b.channels[id]
	b.RUnlock()
	if !cok {
		return
	}

	c.RLock()
	for _, g := range c.gates {
		b.pusher <- message{id: id, addr: g.addr, Message: msg}
	}
	c.RUnlock()
}
