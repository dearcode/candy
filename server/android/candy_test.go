package candy

import (
	"testing"
)

func TestCandy(t *testing.T) {
	c := NewCandy("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	id, err := c.Register("abc", "123")
	if err != nil {
		t.Fatalf("Register error:%s", err.Error())
	}
	t.Logf("register userID:%d", id)
}
