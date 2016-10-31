package gate

import (
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type connection struct {
	user   int64
	addr   string
	device string
	last   time.Time
	stream meta.Gate_StreamServer
}

const (
	connectionLost = time.Minute * 2
)

func newConnection(addr string) *connection {
	return &connection{addr: addr, last: time.Now()}
}

func (c *connection) getAddr() string {
	return c.addr
}

func (c *connection) setUser(id int64) {
	c.user = id
}

func (c *connection) setDevice(dev string) {
	log.Debugf("%s connection device:%v", c.addr, dev)
	c.device = dev
}

func (c *connection) getDevice() string {
	return c.device
}

func (c *connection) waitClose(stream meta.Gate_StreamServer) {
	c.stream = stream
	t := time.NewTicker(util.NetworkTimeout)
	log.Debugf("waitClose add stream from:%s", c.getAddr())
	for {
		<-t.C
		if time.Now().Sub(c.last) > connectionLost {
			log.Debugf("%s timeout, last:%v, now:%v", c.addr, c.last, time.Now())
			stream.Close()
			break
		}
	}
	t.Stop()
}

func (c *connection) getUser() int64 {
	return c.user
}

func (c *connection) heartbeat() {
	log.Debugf("%s connection hearbeat", c.addr)
	c.last = time.Now()
}

func (c *connection) send(msg *meta.PushMessage) error {
	log.Debugf("%s msg:%v", c.addr, msg)
	if c.stream == nil {
		log.Errorf("%s stream is nil", c.addr)
		return nil
	}
	return c.stream.Send(msg)
}
