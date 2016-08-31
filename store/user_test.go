package store

import (
	"os"
	"testing"
	"time"
)

var (
	testUserDBPath = "/tmp/test_user.db"
)

func init() {
	if err := os.RemoveAll(testUserDBPath); err != nil {
		println(err.Error())
	}
}

func TestUserDB(t *testing.T) {
	u := newUserDB(testUserDBPath)
	if err := u.start(); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}
	defer u.db.Close()

	user := "test_user"
	pass := "test_passwd"
	uid := int64(time.Now().UnixNano())

	if err := u.register(user, pass, uid); err != nil {
		t.Fatalf("register error:%s", err.Error())
	}

	id, err := u.auth(user, pass)
	if err != nil {
		t.Fatalf("auth error:%s", err.Error())
	}

	if id != uid {
		t.Fatalf("auth id:%d expect:%d", id, uid)
	}

}

func TestUserLastMessage(t *testing.T) {
	u := newUserDB(testUserDBPath)
	if err := u.start(); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}
	defer u.db.Close()

	uid := int64(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		mid := int64(i * 100)
		if err := u.addMessage(uid, mid); err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}

		last, err := u.getLastMessageID(uid)
		if err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}
		if last != mid {
			t.Fatalf("last:%d, expect:%d", last, mid)
		}
	}
}

func TestUserMessageDB(t *testing.T) {
	u := newUserDB(testUserDBPath)
	if err := u.start(); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}
	defer u.db.Close()

	uid := int64(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		mid := int64(i)
		if err := u.addMessage(uid, mid); err != nil {
			t.Fatalf("add message error:%s", err.Error())
		}
	}

	ids, err := u.getMessageIDs(uid, int64(0))
	if err != nil {
		t.Fatalf("getMessageIDs error:%s", err.Error())
	}

	for i := 0; i < 10; i++ {
		if ids[i] != int64(i) {
			t.Fatalf("ids %d expect:%d, find:%d, ids:%+v", i, i, ids[i], ids)
		}
	}
}
