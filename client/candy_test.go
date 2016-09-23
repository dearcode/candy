package candy

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/dearcode/candy/util/log"
)

var (
	userNames map[string]int64
	passwd    map[string]string
	client    *CandyClient
)

type cmdClient struct {
}

// OnRecv 这函数理论上是多线程调用，客户端需要注意下
func (c *cmdClient) OnRecv(id int64, method int, group int64, from int64, to int64, body string) {
	fmt.Printf("recv msg id:%d method:%d, group:%d, from:%d, to:%d, body:%s\n", id, method, group, from, to, body)
}

// OnError 连接被服务器断开，或其它错误
func (c *cmdClient) OnError(msg string) {
	fmt.Printf("rpc error:%s\n", msg)
}

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

	time.Sleep(time.Second * 3)

	client = NewCandyClient("0.0.0.0:9000", &cmdClient{})
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

		//根据用户名查询用户信息
		userInfo, err := client.GetUserInfoByName(name)
		if err != nil {
			t.Fatalf("get userInfo error:%v, user:%s", err, name)
		}

		if userInfo.NickName != nickName {
			t.Fatalf("nick name not match, user:%s, expect:%s, recv:%s", name, nickName, userInfo.NickName)
		}

		//根据用户ID查询用户信息
		userInfo, err = client.GetUserInfoByID(id)
		if err != nil {
			t.Fatalf("get userInfo by id error:%v, userID:%v", err, id)
		}

		t.Logf("GetUserInfoByName success, id:%v user:%v nickName:%v avatar:%v", userInfo.ID, userInfo.Name, userInfo.NickName, userInfo.Avatar)
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
			ok, err := client.AddFriend(uid, confirm, "ok")
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

func TestHeartbeat(t *testing.T) {
	err := client.Heartbeat()
	if err != nil {
		t.Fatalf("Heartbeat error:%v", err)
	}

	t.Logf("Heartbeat success")
}

func TestCreateGroup(t *testing.T) {
	for i := 0; i < 5; i++ {
		gname := fmt.Sprintf("群组%v_%v", i, time.Now().Unix())
		gid, err := client.CreateGroup(gname)
		if err != nil {
			t.Fatalf("CreateGroup error:%v", err)
		}

		t.Logf("CreateGroup success, gid:%v", gid)
	}

	t.Logf("CreateGroup All success")
}

func TestLoadGroupList(t *testing.T) {
	groupList, err := client.LoadGroupList()
	if err != nil {
		t.Fatalf("LoadGroupList error:%v", err)
	}

	for index, group := range groupList.Groups {
		fmt.Printf("group:%v {ID:%v, Name:%v, Users:%v}\n", index, group.ID, group.Name, group.Users)
	}

	t.Logf("LoadGroupList success")
}

func TestLoadFriendList(t *testing.T) {
	friendList, err := client.LoadFriendList()
	if err != nil {
		t.Fatalf("LoadFriendList error:%v", err)
	}

	for index, user := range friendList.Users {
		fmt.Printf("friend%v  userID:%v\n", index, user)
	}

	t.Logf("LoadFriendList success")
}
