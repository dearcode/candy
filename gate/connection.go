package gate

import (
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type connection struct {
	user   int64
	token  int64
	device string
	last   time.Time
	stream meta.Gate_StreamServer
}

const (
	connectionLost = time.Minute * 2
)

func newConnection(user, token int64, dev string) *connection {
	log.Debugf("new connection user:%d token:%d dev:%s", user, token, dev)
	return &connection{user: user, token: token, device: dev, last: time.Now()}
}

func (c *connection) getToken() int64 {
	return c.token
}

func (c *connection) getDevice() string {
	return c.device
}

func (c *connection) getUser() int64 {
	return c.user
}

func (c *connection) waitClose(stream meta.Gate_StreamServer) {
	c.stream = stream
	t := time.NewTicker(util.NetworkTimeout)
	log.Debugf("wait token:%d timeout", c.token)
	for {
		<-t.C
		if time.Now().Sub(c.last) > connectionLost {
			log.Debugf("%d timeout, last:%v, now:%v", c.token, c.last, time.Now())
			break
		}
	}
	t.Stop()
}

func (c *connection) onHeartbeat() {
	log.Debugf("%d connection hearbeat", c.token)
	c.last = time.Now()
}

func (c *connection) send(msg *meta.PushMessage) error {
	log.Debugf("%d msg:%v", c.token, msg)
	if c.stream == nil {
		log.Errorf("%s stream is nil", c.token)
		return nil
	}
	return c.stream.Send(msg)
}
