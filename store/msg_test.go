package store

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dearcode/candy/meta"
)

type testSender struct {
}

func (s *testSender) send(msg meta.Message) error {
	return nil
}

func TestMessageDB(t *testing.T) {
	m := newMessageDB("/tmp/m.db")
	if err := m.start(&testSender{}); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}

	var msgs []meta.Message
	for i := 0; i < 10; i++ {
		id := time.Now().UnixNano()
		msg := meta.Message{ID: int64(id), Body: fmt.Sprintf("this is msg:%d", id)}
		msgs = append(msgs, msg)
		if err := m.add(msg); err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}
		time.Sleep(time.Millisecond)
	}

	for _, msg := range msgs {
		mss, err := m.get(msg.ID)
		if err != nil {
			t.Fatalf("get msg:(%d) error", msg.ID)
		}

		if !strings.EqualFold(msg.Body, mss[0].Body) {
			t.Fatalf("expect:%s, find:%s", msg.Body, mss[0].Body)
		}
	}
}
