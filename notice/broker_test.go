package notice

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

func TestMain(main *testing.M) {
	debug := flag.Bool("V", false, "set log level:debug")
	flag.Parse()
	if *debug {
		log.SetLevel(log.LOG_DEBUG)
	} else {
		log.SetLevel(log.LOG_ERROR)
	}

	os.Exit(main.Run())
}

type testGateSender struct {
	err error
	c   chan meta.PushRequest
}

func newTestGateSender(err error, c chan meta.PushRequest) *testGateSender {
	return &testGateSender{err: err, c: c}
}

func (t *testGateSender) push(addr string, req meta.PushRequest) error {
	if t.err != nil {
		return t.err
	}
	t.c <- req
	return nil
}

var (
	uid   = int64(1111)
	dev   = "android_xxxxxx"
	token = int64(222222222222)
	c     = make(chan meta.PushRequest)
	host  = "127.0.0.1:9001"
)

func TestBrokerUnSubscribe(t *testing.T) {
	b := newBroker(newTestGateSender(nil, nil))
	//订阅
	b.subscribe(uid, token, dev, host)

	//取消之后应该是找不到这个用户的订阅
	b.unSubscribe(uid, token, dev)
	if _, ok := b.users[uid]; ok {
		t.Fatalf("UnSubscribe user:%d dev:%s error", uid, dev)
	}
}

func TestBrokerSubscribe(t *testing.T) {
	b := newBroker(newTestGateSender(nil, nil))
	//订阅
	b.subscribe(uid, token, dev, host)
	if _, ok := b.users[uid]; !ok {
		t.Fatalf("Subscribe user:%d, result not found", uid)
	}

	devs := b.users[uid]
	if len(devs) != 1 {
		t.Fatalf("Subscribe user:%d dev:%s, find result len:%d", uid, dev, len(devs))
	}

	//订阅两次也不应该有两条结果, 应该是覆盖的
	b.subscribe(uid, token, dev, host)
	devs = b.users[uid]
	if len(devs) != 1 {
		t.Fatalf("ReSubscribe user:%d dev:%s, find result len:%d", uid, dev, len(devs))
	}

}

func TestBrokerPush(t *testing.T) {
	b := newBroker(newTestGateSender(nil, c))
	//订阅
	b.subscribe(uid, token, dev, host)

	msg := meta.PushMessage{Msg: meta.Message{Body: "test"}}
	pushID := meta.PushID{User: uid, Before: 2222}
	b.push(msg, pushID)

	select {
	case nm := <-c:
		t.Logf("recv msg:%v", nm)
	case <-time.After(time.Second):
		log.Fatalf("not found message")
	}
}
