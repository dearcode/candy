package store

import (
	"fmt"
	"testing"

	"github.com/dearcode/candy/server/meta"
)

func TestMessageDB(t *testing.T) {
	m := newMessageDB("/tmp/m.db")
	if err := m.start(); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}

	for i := 0; i < 10; i++ {
		msg := meta.Message{ID: int64(i), Body: fmt.Sprintf("this is msg:%d", i)}
		if err := m.add(msg); err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}
	}

	mss, err := m.getMessageAll()
	if err != nil {
		t.Fatalf("get message error:%s", err.Error())
	}

	for _, ms := range mss {
		fmt.Printf("msg:%+v\n", ms)
	}

	/*
		msg, err := m.getMessage(int64(0))
		if err != nil {
			t.Fatalf("get message error:%s", err.Error())
		}
		fmt.Printf("msg:%+v\n", msg)
	*/
}
