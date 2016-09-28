package notice

import (
	"flag"
	"fmt"
	"os"
	"testing"

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

type testGate struct {
	hosts map[string]chan meta.GateNoticeRequest
}

func newTestGate() *testGate {
	return &testGate{hosts: make(map[string]chan meta.GateNoticeRequest)}
}

func (g *testGate) addHost(addr string, c chan meta.GateNoticeRequest) {
	g.hosts[addr] = c
}

func (g *testGate) notice(addr string, ids []*meta.PushID, msg *meta.Message) error {
	if c, ok := g.hosts[addr]; ok {
		c <- meta.GateNoticeRequest{ID: ids, Msg: msg}
	}
	log.Debugf("notice to %s ids:%v, msg:%v", addr, ids, msg)
	return nil
}

func TestBroker(t *testing.T) {
	g := newTestGate()
	b := newBroker(g)

	b.Start()

	type r struct {
		c chan meta.GateNoticeRequest
		h string
		i int64
	}

	var rs []r
	var ids []*meta.PushID

	for i := 0; i < 10; i++ {
		d := r{c: make(chan meta.GateNoticeRequest), h: fmt.Sprintf("127.0.0.%d:%d", i, i), i: int64(i)}
		rs = append(rs, d)
		g.addHost(d.h, d.c)
		b.Subscribe(d.i, d.h)
		ids = append(ids, &meta.PushID{User: d.i, Before: d.i})
	}

	go func() {
		msg := meta.Message{Group: 1, Body: "test"}
		b.Push(msg, ids...)
	}()

	for i := 0; i < 10; i++ {
		d := <-rs[i].c
		if d.ID[0].User != rs[i].i {
			t.Fatalf("msg[%d] user:%d, expect:%d", i, d.ID[0].User, rs[i].i)
		}
	}
}
