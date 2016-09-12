package candy

import (
	"testing"
)

func TestCandy(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	id, err := c.Register("abcd", "123")
	if err != nil {
		t.Fatalf("Register error:%v", err)
	}
	t.Logf("register userID:%d", id)

	id, err = c.Login("abcd", "123")
	if err != nil {
		t.Fatalf("Login error:%v", err)
	}
	t.Logf("login success, userID:%d", id)
}
