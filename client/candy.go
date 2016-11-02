package client

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	networkTimeout = time.Second * 5
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
	broken  bool
	last    time.Time
	conn    *grpc.ClientConn
	gate    meta.GateClient
	handler MessageHandler
	stream  meta.Gate_StreamClient
	id      int64
	token   int64
	user    string
	pass    string
	device  string
	sync.RWMutex
}

// NewCandyClient - create an new CandyClient
func NewCandyClient(host string, handler MessageHandler) *CandyClient {
	return &CandyClient{host: host, handler: handler, broken: true}
}

// Start 连接服务端.
func (c *CandyClient) Start() error {
	var err error

	c.conn, err = grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		log.Errorf("dial:%s error:%s", c.host, err.Error())
		return err
	}

	c.gate = meta.NewGateClient(c.conn)
	c.last = time.Now()

	go c.healthCheck()

	return nil
}

// service 调用服务器接口, 带上token
func (c *CandyClient) service(call func(context.Context, meta.GateClient) error) {
	ctx := util.ContextSet(context.Background(), "token", fmt.Sprintf("%d", c.token))
	ctx = util.ContextSet(ctx, "id", fmt.Sprintf("%d", c.id))
	if err := call(ctx, c.gate); err != nil {
		log.Infof("call:%s error:%s", c.host, err.Error())
		return
	}
	c.Lock()
	c.last = time.Now()
	c.Unlock()
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
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
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

// Login 用户登陆, 如果发生连接断开，一定要重新登录
func (c *CandyClient) Login(user, passwd string) (int64, error) {
	if code, err := CheckUserName(user); err != nil {
		return -1, NewError(code, err.Error())
	}

	if code, err := CheckUserPassword(passwd); err != nil {
		return -1, NewError(code, err.Error())
	}

	req := &meta.GateUserLoginRequest{User: user, Password: passwd}
	var resp *meta.GateUserLoginResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.Login(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return -1, err
	}

	c.token = resp.Token
	c.id = resp.ID
	c.user = user
	c.pass = passwd

	stream, err := c.openStream()
	if err != nil {
		return -1, err
	}

	go c.receiver(stream)

	return resp.ID, nil
}

// Logout 注销登陆
func (c *CandyClient) Logout() error {
	c.user = ""
	c.pass = ""
	c.token = 0
	c.id = 0

	req := &meta.GateUserLogoutRequest{}
	var resp *meta.GateUserLogoutResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.Logout(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
	var resp *meta.GateUpdateUserInfoResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.UpdateUserInfo(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
	var resp *meta.GateUpdateSignatureResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.UpdateSignature(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
	var resp *meta.GateUpdateUserPasswordResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.UpdateUserPassword(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
		return "", err
	}

	data, err := encodeJSON(userInfo)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *CandyClient) getUserInfoByName(user string) (*meta.UserInfo, error) {
	if code, err := CheckUserName(user); err != nil {
		return nil, NewError(code, err.Error())
	}

	req := &meta.GateGetUserInfoRequest{FindByName: true, UserName: user}
	var resp *meta.GateGetUserInfoResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.GetUserInfo(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
		return "", err
	}

	data, err := encodeJSON(userInfo)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *CandyClient) getUserInfoByID(userID int64) (*meta.UserInfo, error) {
	req := &meta.GateGetUserInfoRequest{UserID: userID}
	var resp *meta.GateGetUserInfoResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.GetUserInfo(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
	var resp *meta.GateFriendResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.Friend(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

// LoadFriendList 加载好友列表
func (c *CandyClient) LoadFriendList() (string, error) {
	req := &meta.GateLoadFriendListRequest{}
	var resp *meta.GateLoadFriendListResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.LoadFriendList(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	friendList := &meta.FriendList{Users: resp.Users}
	data, err := encodeJSON(friendList)
	if err != nil {
		return "", err
	}

	return string(data), resp.Header.Error()
}

// FindUser 支持模糊查询，返回对应用户的列表
func (c *CandyClient) FindUser(user string) (string, error) {
	req := &meta.GateFindUserRequest{User: user}
	var resp *meta.GateFindUserResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.FindUser(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	var users []*meta.UserInfo
	for _, matchUser := range resp.Users {
		userInfo, e := c.getUserInfoByName(matchUser)
		if e != nil {
			return "", e
		}
		users = append(users, userInfo)
	}
	userList := &meta.UserList{Users: users}
	data, err := encodeJSON(userList)
	if err != nil {
		return "", err
	}

	return string(data), resp.Header.Error()
}

// FileExist 判断文件是否存在
func (c *CandyClient) FileExist(key string) (bool, error) {
	req := &meta.GateCheckFileRequest{Names: []string{key}}
	var resp *meta.GateCheckFileResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.CheckFile(ctx, req); err != nil {
			return err
		}
		return nil
	})
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
	var resp *meta.GateUploadFileResponse
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.UploadFile(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return md5, err
	}

	return md5, resp.Header.Error()
}

// FileDownload 文件下载
func (c *CandyClient) FileDownload(key string) ([]byte, error) {
	req := &meta.GateDownloadFileRequest{Names: []string{key}}
	var resp *meta.GateDownloadFileResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.DownloadFile(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp.Files[key], resp.Header.Error()
}

// SendMessage 向服务器发送消息.
func (c *CandyClient) SendMessage(group, to int64, body string) (int64, error) {
	req := &meta.GateSendMessageRequest{Msg: &meta.Message{Group: group, To: to, Body: body}}
	var resp *meta.GateSendMessageResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.SendMessage(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return resp.ID, resp.Header.Error()
}

func (c *CandyClient) openStream() (resp meta.Gate_StreamClient, err error) {
	req := &meta.GateStreamRequest{Token: c.token, ID: c.id}
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.Stream(ctx, req); err != nil {
			return err
		}
		return nil
	})
	return
}

// receiver 一直接收服务器返回消息, 直到出错.
func (c *CandyClient) receiver(stream meta.Gate_StreamClient) {
	for {
		pm, err := stream.Recv()
		if err != nil {
			log.Errorf("recv error:%s", err)
			c.onError(err.Error())
			break
		}
		c.handler.OnRecv(int32(pm.Event), int32(pm.Operate), pm.Msg.ID, pm.Msg.Group, pm.Msg.From, pm.Msg.To, pm.Msg.Body)
	}
}

func (c *CandyClient) onError(msg string) {
	c.Lock()
	c.last = time.Now().Add(-time.Minute)
	if c.broken {
		c.Unlock()
		return
	}
	c.broken = true
	c.Unlock()

	if strings.Contains(msg, "invalid context") && c.user != "" && c.pass != "" {
		c.Login(c.user, c.pass)
	}

	c.handler.OnError(msg)
}

//onHealth 如果网络正常了，要尝试启动Push Stream
func (c *CandyClient) onHealth() {
	c.Lock()
	c.last = time.Now()
	if !c.broken {
		c.Unlock()
		return
	}
	c.broken = false
	c.Unlock()

	c.handler.OnHealth()

	if c.token != 0 && c.id != 0 {
		stream, err := c.openStream()
		if err != nil {
			c.onError(err.Error())
		}

		go c.receiver(stream)

	}

}

// OnNetStateChange 移动端如果网络状态发生变化要通知这边
func (c *CandyClient) OnNetStateChange() {
	//TODO 细分
	c.Lock()
	c.last = time.Now().Add(-time.Minute)
	c.Unlock()
}

// healthCheck 健康检查,60秒发一次, 目前服务器超过90秒会发探活
func (c *CandyClient) healthCheck() {
	t := time.NewTicker(networkTimeout)
	defer t.Stop()

	for {
		<-t.C
		c.RLock()
		if time.Now().Sub(c.last) < time.Minute {
			c.RUnlock()
			continue
		}
		c.RUnlock()

		_, err := healthpb.NewHealthClient(c.conn).Check(context.Background(), &healthpb.HealthCheckRequest{})
		if err != nil {
			log.Errorf("healthCheck error:%v", err)
			c.onError(err.Error())
			continue
		}
		log.Debugf("onHealth")
		c.onHealth()
	}
}

// CreateGroup 创建群组
func (c *CandyClient) CreateGroup(name string) (int64, error) {
	req := &meta.GateGroupCreateRequest{Name: name}
	var resp *meta.GateGroupCreateResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.GroupCreate(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

// Group 群操作
func (c *CandyClient) Group(id int64, operate int32, users []int64, msg string) error {
	req := &meta.GateGroupRequest{ID: id, Msg: msg, Operate: meta.Relation(operate), Users: users}
	var resp *meta.GateGroupResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.Group(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return resp.Header.Error()
}

// DeleteGroup 解散群组
func (c *CandyClient) DeleteGroup(id int64) error {
	req := &meta.GateGroupDeleteRequest{ID: id}
	var resp *meta.GateGroupDeleteResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.GroupDelete(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return resp.Header.Error()
}

// LoadGroupList 拉取群组列表
func (c *CandyClient) LoadGroupList() (string, error) {
	req := &meta.GateLoadGroupListRequest{}
	var resp *meta.GateLoadGroupListResponse
	var err error
	c.service(func(ctx context.Context, api meta.GateClient) error {
		if resp, err = api.LoadGroupList(ctx, req); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	var groups []*meta.GroupInfo
	for _, group := range resp.Groups {
		groups = append(groups, &meta.GroupInfo{ID: group.ID, Name: group.Name, Member: group.Member, Admins: group.Admins})
	}

	groupList := &meta.GroupList{Groups: groups}
	data, err := encodeJSON(groupList)
	if err != nil {
		return "", err
	}

	return string(data), resp.Header.Error()
}
