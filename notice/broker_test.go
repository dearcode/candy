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

func TestBrokerUnSubscribe(t *testing.T) {
	b := newBroker()
	b.start()
	uid := int64(1111)
	dev := "android_xxxxxx"
	c := make(chan pushRequest)
	cid := b.addPushChan(c)
	//订阅
	b.subscribe(uid, dev, cid)

	//取消之后应该是找不到这个用户的订阅
	b.unSubscribe(uid, dev)
	if _, ok := b.users[uid]; ok {
		t.Fatalf("UnSubscribe user:%d dev:%s error", uid, dev)
	}
}

func TestBrokerSubscribe(t *testing.T) {
	b := newBroker()
	b.start()

	uid := int64(1111)
	dev := "android_xxxxxx"
	c := make(chan pushRequest)
	cid := b.addPushChan(c)
	b.subscribe(uid, dev, cid)

	if _, ok := b.users[uid]; !ok {
		t.Fatalf("Subscribe user:%d, result not found", uid)
	}

	devs := b.users[uid]
	if len(devs) != 1 {
		t.Fatalf("Subscribe user:%d dev:%s, find result len:%d", uid, dev, len(devs))
	}

	//订阅两次也不应该有两条结果, 应该是覆盖的
	b.subscribe(uid, dev, cid)

	devs = b.users[uid]
	if len(devs) != 1 {
		t.Fatalf("ReSubscribe user:%d dev:%s, find result len:%d", uid, dev, len(devs))
	}

}

func TestBrokerPush(t *testing.T) {
	b := newBroker()
	b.start()

	uid := int64(1111)
	dev := "android_xxxxxx"
	c := make(chan pushRequest, 1000)

	cid := b.addPushChan(c)
	b.subscribe(uid, dev, cid)

	msg := meta.PushMessage{Msg: &meta.Message{Body: "test"}}
	pushID := meta.PushID{User: uid, Before: 2222}
	b.push(msg, pushID)

	select {
	case nm := <-c:
		t.Logf("recv msg:%v", nm)
	case <-time.After(time.Second):
		log.Fatalf("not found message")
	}
}

func TestBrokerDelPushChan(t *testing.T) {
	b := newBroker()
	b.start()
	c := make(chan pushRequest)
	cid := b.addPushChan(c)

	b.delPushChan(cid)

	if _, ok := b.chans[cid]; ok {
		t.Fatalf("find delete's chan:%d", cid)
	}

}

func TestBrokerAddPushChan(t *testing.T) {
	b := newBroker()
	b.start()
	c := make(chan pushRequest)
	cid := b.addPushChan(c)
	if cid != 1 {
		t.Fatalf("add first chan id must be 1, recv:%d", cid)
	}
	if _, ok := b.chans[cid]; !ok {
		t.Fatalf("chan:%d not found", cid)
	}
}
