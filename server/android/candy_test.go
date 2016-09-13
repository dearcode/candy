package candy

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var (
	userName   = fmt.Sprintf("testuser%v", time.Now().Unix())
	userPasswd = "testpwd"
)

func TestRegister(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	id, err := c.Register(userName, userPasswd)
	if err != nil {
		t.Fatalf("Register error:%v", err)
	}

	t.Logf("register success userID:%d userName:%v userPasswd:%v", id, userName, userPasswd)
}

func TestMultiRegister(t *testing.T) {
	count := rand.Intn(20)
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

	id, err := c.Login(userName, userPasswd)
	if err != nil {
		t.Fatalf("Login error:%v", err)
	}

	t.Logf("login success, userID:%d userName:%v userPasswd:%v", id, userName, userPasswd)
}

func TestUpdateUserInfo(t *testing.T) {
	c := NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		t.Fatalf("start client error:%s", err.Error())
	}

	//first need login
	id, err := c.Login(userName, userPasswd)
	if err != nil {
		t.Fatalf("Login error:%v", err)
	}

	//random nickName
	nickName := fmt.Sprintf("nickName%v", time.Now().Unix())

	id, err = c.UpdateUserInfo(userName, nickName, nil)
	if err != nil {
		t.Fatalf("UpdateUserInfo error:%v", err)
	}

	t.Logf("UpdateUserInfo success, userID:%d userName:%v nickName:%v", id, userName, nickName)

	id, name, nick, avatar, err := c.GetUserInfo(userName)
	if err != nil {
		t.Fatalf("get userInfo error:%v", err)
	}

	if nickName != nick {
		t.Fatalf("nick name not match")
	}

	t.Logf("GetUserInfo success, id:%v user:%v nickName:%v avatar:%v", id, name, nick, avatar)
}
