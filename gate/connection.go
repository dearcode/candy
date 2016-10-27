package gate

import (
	"sync"
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type connection struct {
	user   int64
	addr   string
	device string
	last   int64
	wg     sync.WaitGroup
	stream meta.Gate_StreamServer
}

func newConnection(addr string) *connection {
	return &connection{addr: addr, last: time.Now().Unix()}
}

func (c *connection) getAddr() string {
	log.Debugf("connection getAddr addr:%v", c.addr)
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
	c.wg.Add(1)
	log.Debugf("%s connection wait stream:%v close", c.addr, stream)
	c.stream = stream
	c.wg.Wait()
}

func (c *connection) getUser() int64 {
	log.Debugf("%s connection getUser:%v", c.addr, c.user)
	return c.user
}

func (c *connection) heartbeat() {
	log.Debugf("%s connection hearbeat", c.addr)
	c.last = time.Now().Unix()
}

func (c *connection) send(msg *meta.PushMessage) error {
	log.Debugf("%s connection send msg:%v", c.addr, msg)
	return c.stream.Send(msg)
}

// close TODO 修改grpc支持关连接
func (c *connection) close() {
	c.wg.Done()
}
