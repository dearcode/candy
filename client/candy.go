package client

import (
	"sync"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	networkTimeout = time.Second * 5
	emptyString    = ""
)

// MessageHandler 接收服务器端推送来的消息
type MessageHandler interface {
	// OnRecv 这函数理论上是多线程调用，客户端需要注意下
	OnRecv(event int32, operate int32, ID int64, group int64, from int64, to int64, body string)

	// OnError 连接被服务器断开，或其它错误
	OnError(msg string)

	// OnHealth 连接正常
	OnHealth()

	// OnUnHealth 连接异常
	OnUnHealth(msg string)
}

// CandyClient 客户端提供和服务器交互的接口
type CandyClient struct {
	host    string
	last    time.Time
	conn    *grpc.ClientConn
	api     meta.GateClient
	handler MessageHandler
	stream  meta.Gate_StreamClient
	health  healthpb.HealthClient
	bhealth bool
	sync.RWMutex
}

// NewCandyClient - create an new CandyClient
func NewCandyClient(host string, handler MessageHandler) *CandyClient {
	return &CandyClient{host: host, handler: handler}
}

// Start 连接服务端.
func (c *CandyClient) Start() error {
	conn, err := grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return errors.Trace(err)
	}

	client := meta.NewGateClient(conn)
	stream, err := client.Stream(context.Background(), &meta.Message{})
	if err != nil {
		return errors.Trace(err)
	}

	c.conn = conn
	c.stream = stream
	c.api = client

	go c.loopRecvMessage(stream)
	go c.healthCheck(healthpb.NewHealthClient(conn))

	return nil
}

// service 调用服务器接口，如果调用直接返回error要原样返回，如果是response里的error自己就行了，这里主要处理网络问题
func (c *CandyClient) service(call func(context.Context, meta.GateClient) error) error {
	var err error

	c.Lock()
	defer c.Unlock()

	for r := util.NewRetry(util.RetryDurationMax(networkTimeout)); r.Valid(); r.Next() {
		log.Debugf("retry:%d", r.Attempts())
		ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
		e := call(ctx, c.api)
		cancel()

		if e != nil {
			c.Stop()
			err = e
			if e = c.Start(); e != nil {
				log.Errorf("start error:%s", errors.ErrorStack(e))
				c.handler.OnError(e.Error())
				err = e
				continue
			}
		}
		break
	}

	return err
}

// Register 用户注册接口
func (c *CandyClient) Register(user, passwd string) (int64, error) {
	if code, err := CheckUserName(user); err != nil {
		return -1, NewError(code, err.Error())
	}
	if code, err := CheckUserPassword(passwd); err != nil {
		return -1, NewError(code, err.Error())
	}

	req := &meta.GateRegisterRequest{User: user, Password: passwd}
	var resp *meta.GateRegisterResponse
	err := c.service(func(ctx context.Context, api meta.GateClient) error {
		var err error
		if resp, err = api.Register(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return -1, err
	}
	log.Debugf("resp:%+v", resp)

	return resp.ID, resp.Header.JsonError()
}

// Login 用户登陆
func (c *CandyClient) Login(user, passwd string) (int64, error) {
	if code, err := CheckUserName(user); err != nil {
		return -1, NewError(code, err.Error())
	}

	if code, err := CheckUserPassword(passwd); err != nil {
		return -1, NewError(code, err.Error())
	}

	req := &meta.GateUserLoginRequest{User: user, Password: passwd}
	resp, err := c.api.Login(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.JsonError()
}

// Logout 注销登陆
func (c *CandyClient) Logout() error {
	req := &meta.GateUserLogoutRequest{}
	resp, err := c.api.Logout(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.JsonError()
}

// UpdateUserInfo 更新用户信息， 昵称/头像
func (c *CandyClient) UpdateUserInfo(user, nickName, avatar string) error {
	if code, err := CheckUserName(user); err != nil {
		return NewError(code, err.Error())
	}

	if code, err := CheckNickName(nickName); err != nil {
		return NewError(code, err.Error())
	}

	req := &meta.GateUpdateUserInfoRequest{Name: user, NickName: nickName, Avatar: avatar}
	resp, err := c.api.UpdateUserInfo(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

// UpdateUserSignature 更新用户签名
func (c *CandyClient) UpdateUserSignature(name, signature string) error {
	if code, err := CheckUserName(name); err != nil {
		return NewError(code, err.Error())
	}

	req := &meta.GateUpdateSignatureRequest{Name: name, Signature: signature}
	resp, err := c.api.UpdateSignature(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

// UpdateUserPassword 更新用户密码
func (c *CandyClient) UpdateUserPassword(user, passwd string) error {
	if code, err := CheckUserName(user); err != nil {
		return NewError(code, err.Error())
	}

	if code, err := CheckUserPassword(passwd); err != nil {
		return NewError(code, err.Error())
	}

	req := &meta.GateUpdateUserPasswordRequest{Name: user, Password: passwd}
	resp, err := c.api.UpdateUserPassword(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

// GetUserInfoByName 根据用户名获取用户信息
//TODO 需要把返回字符串修改成对应的类型
func (c *CandyClient) GetUserInfoByName(user string) (string, error) {
	userInfo, err := c.getUserInfoByName(user)
	if err != nil {
		return emptyString, err
	}

	data, err := encodeJSON(userInfo)
	if err != nil {
		return emptyString, err
	}

	return string(data), nil
}

func (c *CandyClient) getUserInfoByName(user string) (*meta.UserInfo, error) {
	if code, err := CheckUserName(user); err != nil {
		return nil, NewError(code, err.Error())
	}

	req := &meta.GateGetUserInfoRequest{FindByName: true, UserName: user}
	resp, err := c.api.GetUserInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}

	userInfo := &meta.UserInfo{
		ID:        resp.ID,
		Name:      resp.User,
		NickName:  resp.NickName,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
	}
	return userInfo, resp.Header.Error()
}

// GetUserInfoByID 根据用户ID获取用户信息
//TODO 需要把返回字符串修改成对应的类型
func (c *CandyClient) GetUserInfoByID(userID int64) (string, error) {
	userInfo, err := c.getUserInfoByID(userID)
	if err != nil {
		return emptyString, err
	}

	data, err := encodeJSON(userInfo)
	if err != nil {
		return emptyString, err
	}

	return string(data), nil
}

func (c *CandyClient) getUserInfoByID(userID int64) (*meta.UserInfo, error) {
	req := &meta.GateGetUserInfoRequest{UserID: userID}
	resp, err := c.api.GetUserInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}

	userInfo := &meta.UserInfo{
		ID:        resp.ID,
		Name:      resp.User,
		NickName:  resp.NickName,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
	}
	return userInfo, resp.Header.Error()
}

// Friend 添加好友
func (c *CandyClient) Friend(userID int64, operate int32, msg string) error {
	req := &meta.GateFriendRequest{UserID: userID, Operate: meta.Relation(operate), Msg: msg}
	resp, err := c.api.Friend(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

// LoadFriendList 加载好友列表
func (c *CandyClient) LoadFriendList() (string, error) {
	req := &meta.GateLoadFriendListRequest{}
	resp, err := c.api.LoadFriendList(context.Background(), req)
	if err != nil {
		return emptyString, err
	}

	friendList := &meta.FriendList{Users: resp.Users}
	data, err := encodeJSON(friendList)
	if err != nil {
		return emptyString, err
	}

	return string(data), resp.Header.Error()
}

// FindUser 支持模糊查询，返回对应用户的列表
func (c *CandyClient) FindUser(user string) (string, error) {
	req := &meta.GateFindUserRequest{User: user}
	resp, err := c.api.FindUser(context.Background(), req)
	if err != nil {
		return emptyString, err
	}

	var users []*meta.UserInfo
	for _, matchUser := range resp.Users {
		userInfo, e := c.getUserInfoByName(matchUser)
		if e != nil {
			return emptyString, e
		}
		users = append(users, userInfo)
	}
	userList := &meta.UserList{Users: users}
	data, err := encodeJSON(userList)
	if err != nil {
		return emptyString, err
	}

	return string(data), resp.Header.Error()
}

// FileExist 判断文件是否存在
func (c *CandyClient) FileExist(key string) (bool, error) {
	req := &meta.GateCheckFileRequest{Names: []string{key}}
	resp, err := c.api.CheckFile(context.Background(), req)
	if err != nil {
		return false, err
	}

	if err = resp.Header.Error(); err != nil {
		return false, err
	}

	if len(resp.Names) == 0 {
		return true, nil
	}

	return false, resp.Header.Error()
}

// FileUpload 文件上传
func (c *CandyClient) FileUpload(data []byte) (string, error) {
	md5 := string(util.MD5(data))
	exist, err := c.FileExist(md5)
	if err != nil {
		return md5, err
	}
	//已有别人上传过了
	if exist {
		return md5, nil
	}

	req := &meta.GateUploadFileRequest{File: data}
	resp, err := c.api.UploadFile(context.Background(), req)
	if err != nil {
		return md5, err
	}

	return md5, resp.Header.Error()
}

// FileDownload 文件下载
func (c *CandyClient) FileDownload(key string) ([]byte, error) {
	req := &meta.GateDownloadFileRequest{Names: []string{key}}
	resp, err := c.api.DownloadFile(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Files[key], resp.Header.Error()
}

// SendMessage 向服务器发送消息.
func (c *CandyClient) SendMessage(group, to int64, body string) (int64, error) {
	req := &meta.GateSendMessageRequest{Msg: &meta.Message{Group: group, To: to, Body: body}}
	resp, err := c.api.SendMessage(context.Background(), req)
	if err != nil {
		return 0, err
	}
	return resp.ID, resp.Header.Error()
}

// Stop 停止接收消息及心跳
func (c *CandyClient) Stop() {
	c.conn.Close()
	log.Debugf("close conn")
}

// loopRecvMessage 一直接收服务器返回消息, 直到退出.
func (c *CandyClient) loopRecvMessage(stream meta.Gate_StreamClient) {
	for {
		pm, err := stream.Recv()
		if err != nil {
			log.Errorf("loopRecvMessage error:%s", errors.ErrorStack(err))
			c.handler.OnError(err.Error())
			break
		}
		c.handler.OnRecv(int32(pm.Event), int32(pm.Operate), pm.Msg.ID, pm.Msg.Group, pm.Msg.From, pm.Msg.To, pm.Msg.Body)
	}
	c.Stop()
}

// healthCheck 健康检查
func (c *CandyClient) healthCheck(health healthpb.HealthClient) {
	t := time.NewTicker(time.Second)
	for {
		<-t.C
		_, err := health.Check(context.Background(), &healthpb.HealthCheckRequest{})
		if err != nil {
			log.Errorf("healthCheck error:%s", errors.ErrorStack(err))
			c.handler.OnError(err.Error())
			break
		}
	}
	t.Stop()
	c.Stop()
}

// CreateGroup 创建群组
func (c *CandyClient) CreateGroup(name string) (int64, error) {
	req := &meta.GateGroupCreateRequest{Name: name}
	resp, err := c.api.GroupCreate(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

// Group 群操作
func (c *CandyClient) Group(id int64, operate int32, users []int64, msg string) error {
	req := &meta.GateGroupRequest{ID: id, Msg: msg, Operate: meta.Relation(operate), Users: users}
	resp, err := c.api.Group(context.Background(), req)
	if err != nil {
		return err
	}
	return resp.Header.Error()
}

// DeleteGroup 解散群组
func (c *CandyClient) DeleteGroup(id int64) error {
	req := &meta.GateGroupDeleteRequest{ID: id}
	resp, err := c.api.GroupDelete(context.Background(), req)
	if err != nil {
		return err
	}
	return resp.Header.Error()
}

// LoadGroupList 拉取群组列表
func (c *CandyClient) LoadGroupList() (string, error) {
	req := &meta.GateLoadGroupListRequest{}
	resp, err := c.api.LoadGroupList(context.Background(), req)
	if err != nil {
		return emptyString, err
	}

	var groups []*meta.GroupInfo
	for _, group := range resp.Groups {
		groups = append(groups, &meta.GroupInfo{ID: group.ID, Name: group.Name, Member: group.Member, Admins: group.Admins})
	}

	groupList := &meta.GroupList{Groups: groups}
	data, err := encodeJSON(groupList)
	if err != nil {
		return emptyString, err
	}

	return string(data), resp.Header.Error()
}
