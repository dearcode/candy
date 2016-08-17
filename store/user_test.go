package store

import (
	"testing"
)

func TestDB(t *testing.T) {
	u := newUserDB("/tmp/u.db")
	if err := u.start(); err != nil {
		t.Fatalf("start error:%s", err.Error())
	}

	user := "test_user"
	pass := "test_passwd"
	uid := int64(11111)

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
