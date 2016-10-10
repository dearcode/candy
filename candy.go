package main

/*
命令行客户端
用于基本功能校验
*/

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	candy "github.com/dearcode/candy/client"
	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

func notice() {
	help := `-----------------------------------------------
1. 注册用户
2. 登陆
3. 注销
4. 更新用户信息
5. 获取用户信息
6. 查找用户
7. 添加好友
8. 发送消息
9. 创建群组
10. 加载群组
11. 根据用户ID获取用户信息
12. 加载好友列表
13. 更改用户密码
14. 确认添加好友
0. 退出
-----------------------------------------------`
	fmt.Println(help)
}

func endSection() {
	fmt.Println("-----------------------------------------------")
}

func register(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------用户注册-----------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	fmt.Println("请输入密码:")
	data, _, _ = reader.ReadLine()
	userPassword := string(data)

	id, err := c.Register(userName, userPassword)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Register code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("Register success, userID:%v userName:%v userPassword:%v", id, userName, userPassword)
}

func login(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------用户登陆-----------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	fmt.Println("请输入密码:")
	data, _, _ = reader.ReadLine()
	userPassword := string(data)

	id, err := c.Login(userName, userPassword)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Login code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("Login success, userID:%v userName:%v userPassword:%v", id, userName, userPassword)
}

func logout(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------注销---------------------------")
	defer endSection()

	err := c.Logout()
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Logout code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("Logout success")
}

func updateUserInfo(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------更新用户信息-------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	fmt.Println("请输入用户昵称：")
	data, _, _ = reader.ReadLine()
	nickName := string(data)

	id, err := c.UpdateUserInfo(userName, nickName, nil)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("updateUserInfo code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("updateUserInfo success, userName:%v nickName:%v userID:%v", userName, nickName, id)
}

func getUserInfoByName(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------获取用户信息-------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	d, _, _ := reader.ReadLine()
	userName := string(d)

	data, err := c.GetUserInfoByName(userName)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("getUserInfo code:%v error:%v", e.Code, e.Msg)
		return
	}

	user, err := candy.DecodeUserInfo([]byte(data))
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Decode UserInfo code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("getUserInfo success, userName:%v", userName)
	log.Debugf("user detail, ID:%v Name:%v NickName:%v Avatar:%v", user.ID, user.Name, user.NickName, user.Avatar)
}

func getUserInfoByID(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------获取用户信息-------------------")
	defer endSection()

	fmt.Println("请输入用户ID:")
	d, _, _ := reader.ReadLine()
	userID := string(d)

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		log.Errorf("Parse int error:%v", err)
		return
	}

	data, err := c.GetUserInfoByID(id)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("getUserInfoByID code:%v error:%v", e.Code, e.Msg)
		return
	}

	user, err := candy.DecodeUserInfo([]byte(data))
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Decode UserInfo code:%v error:%v", e.Code, e.Msg)
		return
	}
	log.Debugf("getUserInfoByID success, userID:%v", id)
	log.Debugf("user detail, ID:%v Name:%v NickName:%v Avatar:%v", user.ID, user.Name, user.NickName, user.Avatar)
}

func findUser(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------查找用户-----------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	d, _, _ := reader.ReadLine()
	userName := string(d)

	data, err := c.FindUser(userName)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("findUser code:%v error:%v", e.Code, e.Msg)
		return
	}

	userList, err := candy.DecodeUserList([]byte(data))
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Decode UserList code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("findUser success, userName:%v*", userName)
	for index, user := range userList.Users {
		log.Debugf("user:%d detail, ID:%v Name:%v NickName:%v Avatar:%v", index, user.ID, user.Name, user.NickName, user.Avatar)
	}
}

func addFriend(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------添加好友-----------------------")
	defer endSection()

	fmt.Println("请输入用户ID:")
	data, _, _ := reader.ReadLine()
	userID := string(data)

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	fmt.Println("请输入附加消息:")
	data, _, _ = reader.ReadLine()
	msg := string(data)

	if err = c.Friend(id, int32(meta.Relation_Add), msg); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("addFriend code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("addFriend userID:%v", userID)
}

func confirmFriend(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------确认添加好友-----------------------")
	defer endSection()

	fmt.Println("请输入用户ID:")
	data, _, _ := reader.ReadLine()
	userID := string(data)

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Friend(id, int32(meta.Relation_Confirm), ""); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("confirmFriend code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("confirmFriend userID:%v", userID)
}
func newMessage(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------发送消息-----------------------")
	defer endSection()
	id := int64(0)

	fmt.Println("请输入接收用户ID:")
	data, _, _ := reader.ReadLine()
	userID := string(data)

	user, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	for {
		fmt.Println("请输入消息内容(quit退出):")
		data, _, _ = reader.ReadLine()
		msg := string(data)

		if msg == "quit" {
			break
		}

		id, err = c.SendMessage(0, user, msg)
		if err != nil {
			e := candy.ErrorParse(err.Error())
			log.Errorf("send message code:%v error:%v", e.Code, e.Msg)
			return
		}

		log.Debugf("send msg[%d] success, userID:%v", id, userID)
	}
}

func createGroup(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("----------------创建群组-----------------------")
	defer endSection()

	fmt.Println("请输入群组名称:")
	data, _, _ := reader.ReadLine()
	groupName := string(data)

	gid, err := c.CreateGroup(groupName)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("createGroup code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("createGroup success, groupName:%v groupID:%v", groupName, gid)
}

func loadGroupList(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("---------------加载群组列表--------------------")
	defer endSection()

	data, err := c.LoadGroupList()
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("loadGroupList code:%v error:%v", e.Code, e.Msg)
		return
	}

	gList, err := candy.DecodeGroupList([]byte(data))
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Decode GroupList code:%v error:%v", e.Code, e.Msg)
		return
	}

	for index, group := range gList.Groups {
		log.Debugf("group%v {ID:%v Name:%v Users:%v}", index, group.ID, group.Name, group.Member)
	}

	log.Debugf("loadGroupList success")
}

func loadFriendList(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("---------------加载好友列表--------------------")
	defer endSection()

	data, err := c.LoadFriendList()
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("loadGroupList code:%v error:%v", e.Code, e.Msg)
		return
	}

	fList, err := candy.DecodeFriendList([]byte(data))
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Decode FriendList code:%v error:%v", e.Code, e.Msg)
		return
	}

	for index, user := range fList.Users {
		log.Debugf("friend%v {ID:%v}", index, user)
	}

	log.Debugf("loadFriendList success")
}

func updateUserPasswd(c *candy.CandyClient, reader *bufio.Reader) {
	fmt.Println("---------------更改用户密码--------------------")
	defer endSection()

	fmt.Println("请输入用户名:")
	data, _, _ := reader.ReadLine()
	user := string(data)

	fmt.Println("请输入新密码:")
	data, _, _ = reader.ReadLine()
	pwd := string(data)

	id, err := c.UpdateUserPassword(user, pwd)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("UpdateUserPassword code:%v error:%v", e.Code, e.Msg)
		return
	}
	log.Debugf("UpdateUserPassword success, id:%v", id)
}

type cmdClient struct{}

// OnRecv 这函数理论上是多线程调用，客户端需要注意下
func (c *cmdClient) OnRecv(event int32, operate int32, id int64, group int64, from int64, to int64, body string) {
	log.Debugf("recv msg id:%d event:%v, operate:%v, group:%d, from:%d, to:%d, body:%s\n", id, meta.Event(event), meta.Relation(operate), group, from, to, body)
}

// OnError 连接被服务器断开，或其它错误
func (c *cmdClient) OnError(msg string) {
	log.Errorf("rpc error:%s\n", msg)
}

// OnHealth 连接恢复
func (c *cmdClient) OnHealth() {
	log.Debugf("connection recovery\n")
}

// OnUnHealth 连接异常
func (c *cmdClient) OnUnHealth(msg string) {
	log.Errorf("connection UnHealth, msg:%v\n", msg)
}

func main() {
	c := candy.NewCandyClient("127.0.0.1:9000", &cmdClient{})
	if err := c.Start(); err != nil {
		log.Errorf("start client error:%s", err.Error())
		return
	}

	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		notice()
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "" {
			continue
		}

		log.Debugf("command:%v", command)
		if command == "0" {
			running = false
		} else if command == "1" {
			register(c, reader)
		} else if command == "2" {
			login(c, reader)
		} else if command == "3" {
			logout(c, reader)
		} else if command == "4" {
			updateUserInfo(c, reader)
		} else if command == "5" {
			getUserInfoByName(c, reader)
		} else if command == "6" {
			findUser(c, reader)
		} else if command == "7" {
			addFriend(c, reader)
		} else if command == "8" {
			newMessage(c, reader)
		} else if command == "9" {
			createGroup(c, reader)
		} else if command == "10" {
			loadGroupList(c, reader)
		} else if command == "11" {
			getUserInfoByID(c, reader)
		} else if command == "12" {
			loadFriendList(c, reader)
		} else if command == "13" {
			updateUserPasswd(c, reader)
		} else if command == "14" {
			confirmFriend(c, reader)
		} else {
			log.Errorf("未知命令")
		}
	}
}
