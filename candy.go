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

	"github.com/fatih/color"

	candy "github.com/dearcode/candy/client"
	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

const (
	CmdExit = iota
	CmdRegister

	CmdLogin
	CmdLogout

	CmdUpdateUserInfo
	CmdChangePassword
	CmdUpdateSignature

	CmdFindUser
	CmdGetUserInfoByName
	CmdGetUserInfoByID

	CmdLoadFriendList
	CmdFriendAdd
	CmdFriendAccept
	CmdFriendDel

	CmdSendMessage

	CmdLoadGroupList
	CmdGroupCreate
	CmdGroupInvite
	CmdGroupApply
	CmdGroupAgree
	CmdGroupAccept
	CmdGroupKick
	CmdGroupExit
	CmdGroupDelete
)

type Cmd struct {
	id    int
	title string
}

func notice() {
	cmdList := []Cmd{
		{CmdRegister, "注册用户"},
		{CmdLogin, "登  陆"},
		{CmdLogout, "注  销"},
		{CmdUpdateUserInfo, "修改用户信息"},
		{CmdChangePassword, "修改密码"},
		{CmdUpdateSignature, "修改签名"},

		{CmdFindUser, "查找用户"},
		{CmdGetUserInfoByName, "获取用户信息(Name)"},
		{CmdGetUserInfoByID, "获取用户信息(ID)"},

		{CmdLoadFriendList, "加载好友列表"},
		{CmdFriendAdd, "添加好友"},
		{CmdFriendAccept, "接受好友请求"},
		{CmdFriendDel, "删除好友"},

		{CmdSendMessage, "发送消息"},

		{CmdLoadGroupList, "加载群列表"},
		{CmdGroupCreate, "创建群"},
		{CmdGroupInvite, "邀请用户入群"},
		{CmdGroupApply, "申请入群"},
		{CmdGroupAgree, "同意用户进群"},
		{CmdGroupAccept, "接受群邀请"},
		{CmdGroupKick, "踢用户出群"},
		{CmdGroupExit, "退出群"},
		{CmdGroupDelete, "解散群"},
		{CmdExit, "退出"},
	}

	startSection("操作列表")
	for _, cmd := range cmdList {
		fmt.Printf("\t\t%2d %s\n", cmd.id, cmd.title)
	}
	endSection()

}

func startSection(key string) {
	color.Green("-----------------%s--------------------", key)
}
func endSection() {
	color.Green("-----------------------------------------------")
}

func register(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("用户注册")
	defer endSection()

	color.Yellow("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	color.Yellow("请输入密码:")
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
	startSection("用户登陆")
	defer endSection()

	color.Yellow("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	color.Yellow("请输入密码:")
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
	startSection("注  销")
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
	startSection("更新用户信息")
	defer endSection()

	color.Yellow("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	color.Yellow("请输入用户昵称：")
	data, _, _ = reader.ReadLine()
	nickName := string(data)

	if err := c.UpdateUserInfo(userName, nickName, ""); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("updateUserInfo code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("updateUserInfo success, userName:%v nickName:%v", userName, nickName)
}

func updateSignature(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("更新用户签名")
	defer endSection()

	color.Yellow("请输入用户名:")
	data, _, _ := reader.ReadLine()
	userName := string(data)
	color.Yellow("请输入签名：")
	data, _, _ = reader.ReadLine()
	signature := string(data)

	if err := c.UpdateUserSignature(userName, signature); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("updateUserSignature code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("updateUserSignature success, userName:%v Signature:%v", userName, signature)
}

func getUserInfoByName(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("获取用户信息")
	defer endSection()

	color.Yellow("请输入用户名:")
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
	log.Debugf("user detail, ID:%v Name:%v NickName:%v Avatar:%v Signature:%v",
		user.ID, user.Name, user.NickName, user.Avatar, user.Signature)
}

func getUserInfoByID(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("获取用户信息")
	defer endSection()

	color.Yellow("请输入用户ID:")
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
	log.Debugf("user detail, ID:%v Name:%v NickName:%v Avatar:%v Signature:%v",
		user.ID, user.Name, user.NickName, user.Avatar, user.Signature)
}

func findUser(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("查找用户")
	defer endSection()

	color.Yellow("请输入用户名:")
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
	startSection("添加好友")
	defer endSection()

	color.Yellow("请输入用户ID:")
	data, _, _ := reader.ReadLine()
	userID := string(data)

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	color.Yellow("请输入附加消息:")
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
	startSection("同意添加好友")
	defer endSection()

	color.Yellow("请输入用户ID:")
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
	startSection("发送消息")
	defer endSection()
	id := int64(0)

	color.Yellow("请输入接收用户ID:")
	data, _, _ := reader.ReadLine()
	userID := string(data)

	user, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	for {
		color.Yellow("请输入消息内容(quit退出):")
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
	startSection("创建群组")
	defer endSection()

	color.Yellow("请输入群组名称:")
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

func readID(reader *bufio.Reader) (int64, error) {
	data, _, err := reader.ReadLine()
	if err != nil {
		return -1, err
	}
	return strconv.ParseInt(string(data), 10, 64)
}

func groupInviteUser(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("邀请用户入群")
	defer endSection()

	var gid, uid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	color.Yellow("请输入用户ID:")
	if uid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read uid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Add), []int64{uid}, "邀请你加群"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group invite success, group:%v, user:%v", gid, uid)
}

func groupUserApply(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("申请入群")
	defer endSection()

	var gid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Add), nil, "我要加你的群"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group apply success, group:%v", gid)
}

func groupAgreeUser(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("同意用户入群")
	defer endSection()

	var gid, uid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	color.Yellow("请输入用户ID:")
	if uid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read uid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Confirm), []int64{uid}, "同意你入群了"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group agree success, group:%v, user:%v", gid, uid)
}

func groupUserAccept(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("接受入群邀请")
	defer endSection()

	var gid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Confirm), nil, "我同意进你的群了"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group accept success, group:%v", gid)
}

func groupKickUser(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("踢用户出群")
	defer endSection()

	var gid, uid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	color.Yellow("请输入用户ID:")
	if uid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read uid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Del), []int64{uid}, "你被踢出群了"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group:%v kick user:%v success", gid, uid)
}

func groupUserExit(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("退出群")
	defer endSection()

	var gid int64
	var err error

	color.Yellow("请输入群组ID:")
	if gid, err = readID(reader); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("read gid code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.Group(gid, int32(meta.Relation_Del), nil, "我退群了"); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("group code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group exit success, group:%v", gid)
}

func groupDelete(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("解散群组")
	defer endSection()

	color.Yellow("请输入群组ID:")
	data, _, _ := reader.ReadLine()
	groupID := string(data)

	id, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("Parse int code:%v error:%v", e.Code, e.Msg)
		return
	}

	if err = c.DeleteGroup(id); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("deleteGroup code:%v error:%v", e.Code, e.Msg)
		return
	}

	log.Debugf("group:%v delete success", groupID)
}

func loadGroupList(c *candy.CandyClient, reader *bufio.Reader) {
	startSection("加载群组列表")
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
	startSection("加载好友列表")
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
	startSection("更改用户密码")
	defer endSection()

	color.Yellow("请输入用户名:")
	data, _, _ := reader.ReadLine()
	user := string(data)

	color.Yellow("请输入新密码:")
	data, _, _ = reader.ReadLine()
	pwd := string(data)

	if err := c.UpdateUserPassword(user, pwd); err != nil {
		e := candy.ErrorParse(err.Error())
		log.Errorf("UpdateUserPassword code:%v error:%v", e.Code, e.Msg)
		return
	}
	log.Debugf("UpdateUserPassword success")
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
	//c := candy.NewCandyClient("candy.dearcode.net:9000", &cmdClient{})
	if err := c.Start(); err != nil {
		log.Errorf("start client error:%s", err.Error())
		return
	}

	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		notice()
		id, err := readID(reader)
		if err != nil {
			log.Errorf("read error:%s", err.Error())
			continue
		}
		switch id {
		case CmdExit:
			running = false
		case CmdRegister:
			register(c, reader)
		case CmdLogin:
			login(c, reader)
		case CmdLogout:
			logout(c, reader)

		case CmdUpdateUserInfo:
			updateUserInfo(c, reader)
		case CmdChangePassword:
			updateUserPasswd(c, reader)
		case CmdUpdateSignature:
			updateSignature(c, reader)

		case CmdGetUserInfoByName:
			getUserInfoByName(c, reader)
		case CmdGetUserInfoByID:
			getUserInfoByID(c, reader)

		case CmdFindUser:
			findUser(c, reader)

		case CmdLoadFriendList:
			loadFriendList(c, reader)
		case CmdFriendAdd:
			addFriend(c, reader)
		case CmdFriendAccept:
			confirmFriend(c, reader)
		case CmdFriendDel:

		case CmdSendMessage:
			newMessage(c, reader)

		case CmdLoadGroupList:
			loadGroupList(c, reader)
		case CmdGroupCreate:
			createGroup(c, reader)
		case CmdGroupDelete:
			groupDelete(c, reader)

		case CmdGroupInvite:
			groupInviteUser(c, reader)
		case CmdGroupAccept:
			groupUserAccept(c, reader)
		case CmdGroupAgree:
			groupAgreeUser(c, reader)
		case CmdGroupApply:
			groupUserApply(c, reader)

		case CmdGroupKick:
			groupKickUser(c, reader)
		case CmdGroupExit:
			groupUserExit(c, reader)

		default:
			log.Errorf("未知命令")
		}
	}
}
