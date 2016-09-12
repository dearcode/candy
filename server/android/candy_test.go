package candy

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCandy(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	id, err := c.Register("adlkje", "123")
	if err != nil {
		t.Fatalf("Register error:%v", err)
	}
	t.Logf("register userID:%d", id)

	id, err = c.Login("adlkje", "123")
	if err != nil {
		t.Fatalf("Login error:%v", err)
	}
	t.Logf("login success, userID:%d", id)
}

func TestRegister(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	userName := fmt.Sprintf("testuser%v", time.Now().Unix())
	userPasswd := "testpwd"

	id, err := c.Register(userName, userPasswd)
	if err != nil {
		t.Fatalf("Register error:%v", err)
	}

	t.Logf("register success userID:%d userName:%v userPasswd:%v", id, userName, userPasswd)
}

func TestMultiRegister(t *testing.T) {
	count := rand.Intn(100)
	if count == 0 {
		count = 1
	}

	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	for i := 0; i < count; i++ {
		userName := fmt.Sprintf("testuser%v", time.Now().UnixNano())
		userPasswd := "testpwd"

		id, err := c.Register(userName, userPasswd)
		if err != nil {
			t.Fatalf("Register error:%v", err)
		}

		t.Logf("register %d account success, userID:%v userName:%v userPasswd:%v", i+1, id, userName, userPasswd)
	}

	t.Logf("multi register success, count:%v", count)
}

func TestLogin(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	userName := "testuserxxxxx"
	userPasswd := "testpwd"
	id, err := c.Login(userName, userPasswd)
	if err != nil {
		t.Fatalf("Login error:%v", err)
	}

	t.Logf("login success, userID:%d userName:%v userPasswd:%v", id, userName, userPasswd)
}
