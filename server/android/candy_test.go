package candy

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/dearcode/candy/server/util/log"
)

var (
	userNames map[string]int64
	passwd    map[string]string
	client    *CandyClient
)

func TestMain(main *testing.M) {
	userNames = make(map[string]int64)
	passwd = make(map[string]string)

	debug := flag.Bool("V", false, "set log level:debug")
	flag.Parse()
	if *debug {
		log.SetLevel(log.LOG_DEBUG)
	} else {
		log.SetLevel(log.LOG_ERROR)
	}

	exes := []string{
		"../bin/master",
		"../bin/notice",
		"../bin/store",
		"../bin/gate",
	}

	cmds := []*exec.Cmd{}

	for _, exe := range exes {
		cmd := exec.Command(exe)
		if err := cmd.Start(); err != nil {
			panic(err.Error())
		}
		cmds = append(cmds, cmd)
	}

	time.Sleep(time.Second)

	client = NewCandyClient("0.0.0.0:9000")
	if err := client.Start(); err != nil {
		panic(err.Error())
	}

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("testuser_%v_%d", time.Now().Unix(), i)
		pass := fmt.Sprintf("testpass_%v_%d", time.Now().Unix(), i)
		id, err := client.Register(name, pass)
		if err != nil {
			for _, cmd := range cmds {
				cmd.Process.Kill()
			}
			panic("Register error:" + err.Error())
		}
		userNames[name] = id
		passwd[name] = pass
	}

	ret := main.Run()

	for _, cmd := range cmds {
		cmd.Process.Kill()
	}

	os.Exit(ret)
}

func TestLogin(t *testing.T) {
	for name, id := range userNames {
		uid, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}
		if uid != id {
			t.Fatalf("Login user:%s, expect id:%d, recv id:%d", name, id, uid)
		}
		t.Logf("login success, userID:%d userName:%v userPasswd:%v", uid, name, name)
	}

}

func TestUpdateUserInfo(t *testing.T) {
	for name := range userNames {
		//first need login
		id, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}

		//random nickName
		nickName := fmt.Sprintf("nickName%v", time.Now().Unix())

		if id, err = client.UpdateUserInfo(name, nickName, nil); err != nil {
			t.Fatalf("UpdateUserInfo error:%v, user:%s, nickName:%s", err, name, nickName)
		}

		t.Logf("UpdateUserInfo success, userID:%d userName:%v nickName:%v", id, name, nickName)

		userInfo, err := client.GetUserInfo(name)
		if err != nil {
			t.Fatalf("get userInfo error:%v, user:%s", err, name)
		}

		if userInfo.NickName != nickName {
			t.Fatalf("nick name not match, user:%s, expect:%s, recv:%s", name, nickName, userInfo.NickName)
		}

		t.Logf("GetUserInfo success, id:%v user:%v nickName:%v avatar:%v", userInfo.ID, userInfo.Name, userInfo.NickName, userInfo.Avatar)
	}
}

func TestUpdateUserPassword(t *testing.T) {
	for name := range userNames {
		//first need login
		id, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}

		//random passwd
		newPasswd := fmt.Sprintf("newpwd%v", time.Now().Unix())

		id, err = client.UpdateUserPassword(name, newPasswd)
		if err != nil {
			t.Fatalf("UpdateUserPassword error:%v", err)
		}

		t.Logf("UpdateUserPassword success, userID:%d userName:%v", id, name)

		//Logout
		err = client.Logout(name)
		if err != nil {
			t.Fatalf("user Logout error:%v", err)
		}

		//Login
		id, err = client.Login(name, newPasswd)
		if err != nil {
			t.Fatalf("use new password login err:%v", err)
		}
		t.Logf("test logout success")

		passwd[name] = newPasswd

		t.Logf("UpdateUserPassword success, userID:%d userName:%v, newPasswd:%s", id, name, newPasswd)
	}
}

func TestFindUser(t *testing.T) {
	for name := range userNames {
		//first need login
		id, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}

		t.Logf("Login success id:%v", id)

		for u := range userNames {
			//find user
			users, err := client.FindUser(u)
			if err != nil {
				t.Fatalf("Find user:%s error:%v", u, err)
			}

			if users == nil || len(users.Users) <= 0 {
				t.Fatalf("Find user error, want large than 0")
			}

			t.Logf("Find user:%s success", u)
		}
	}
}

func TestAddFriend(t *testing.T) {
	relation := make(map[int64]int64)
	for name := range userNames {
		id, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}

		for _, uid := range userNames {
			if uid == id {
				//自己不能添加自己
				continue
			}
			//add friend
			confirm := false
			if _, ok := relation[uid]; ok {
				confirm = true
			}
			ok, err := client.AddFriend(uid, confirm)
			if err != nil {
				t.Fatalf("AddFriend error:%v", err)
			}
			// 如果双方都加过好友，这里ok应该返回true
			if relation[uid] == id {
				if !ok {
					t.Fatal("expect ok is true")
				}
			}
			relation[id] = uid
		}
	}
}

func TestFileUploadAndDownload(t *testing.T) {
	keys := make(map[string]struct{})
	for name := range userNames {
		id, err := client.Login(name, passwd[name])
		if err != nil {
			t.Fatalf("Login error:%v", err)
		}
		t.Logf("login user:%v, id:%d", name, id)
		key, err := client.FileUpload([]byte(name))
		if err != nil {
			t.Fatalf("FileUpload error:%v", err)
		}

		t.Logf("upload user:%s, file:%s", name, key)
		if _, ok := keys[key]; ok {
			t.Fatalf("key:%s exist", key)
		}
		keys[key] = struct{}{}

		ok, err := client.FileExist(key)
		if err != nil {
			t.Fatalf("FileExist error:%v", err)
		}

		if !ok {
			t.Fatalf("key:%s not exist", key)
		}
		data, err := client.FileDownload(key)
		if err != nil {
			t.Fatalf("FileDownload error:%v", err)
		}
		if !bytes.Equal(data, []byte(name)) {
			t.Fatalf("FileDownload key:%s, val:%s, expect:%s", key, data, name)
		}
	}
}
