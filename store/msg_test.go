package store

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
)

const (
	testMessageDBPath = "/tmp/test_candy_db"
)

func init() {
	if err := os.RemoveAll(testMessageDBPath); err != nil {
		println(err.Error())
	}
}

type testSender struct {
}

func (s *testSender) send(msg meta.PushMessage) error {
	return nil
}

func TestMessageDB(t *testing.T) {
	m := newMessageDB(testMessageDBPath)
	if err := m.start(&testSender{}); err != nil {
		t.Fatalf("start error:%s", errors.ErrorStack(err))
	}

	var pms []meta.PushMessage
	for i := 0; i < 10; i++ {
		id := time.Now().UnixNano()
		msg := meta.PushMessage{Msg: meta.Message{ID: int64(id), Body: fmt.Sprintf("this is msg:%d", id)}}
		pms = append(pms, msg)
		if err := m.send(msg); err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}
		time.Sleep(time.Millisecond)
	}

	for _, pm := range pms {
		msg, err := m.get(pm.Msg.ID)
		if err != nil {
			t.Fatalf("get msg:(%d) error", pm.Msg.ID)
		}

		if !strings.EqualFold(pm.Msg.Body, msg[0].Msg.Body) {
			t.Fatalf("expect:%s, find:%s", pm.Msg.Body, msg[0].Msg.Body)
		}
	}
}
