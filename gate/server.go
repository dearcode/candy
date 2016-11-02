package gate

import (
	"net"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

var (
	// ErrUndefineMethod 方法未定义.
	ErrUndefineMethod = errors.New("undefine method")
	// ErrInvalidState 当前用户离线或未登录.
	ErrInvalidState = errors.New("invalid context")
)

// Gate recv client request.
type Gate struct {
	store        *storeClient
	manager      *manager
	master       *util.MasterClient
	healthServer *health.Server // nil means disabled
	server       *grpc.Server
}

// NewGate new gate server.
func NewGate(host, master, notifer, store string) (*Gate, error) {
	l, err := net.Listen("tcp", host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	nc, err := util.NewNotiferClient(notifer)
	if err != nil {
		return nil, errors.Trace(err)
	}

	mc, err := util.NewMasterClient(master, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	sc, err := newStoreClient(store)
	if err != nil {
		return nil, errors.Trace(err)
	}

	g := &Gate{
		manager:      newManager(nc, host),
		store:        sc,
		master:       mc,
		healthServer: health.NewServer(),
		server:       grpc.NewServer(),
	}
	meta.RegisterGateServer(g.server, g)

	healthpb.RegisterHealthServer(g.server, g)

	return g, g.server.Serve(l)
}

// Check 心跳
func (g *Gate) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if c := g.manager.getConnection(ctx); c != nil {
		//更新心跳信息
		c.onHeartbeat()
	}
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

// Register user, passwd.
func (g *Gate) Register(ctx context.Context, req *meta.GateRegisterRequest) (*meta.GateRegisterResponse, error) {
	id, err := g.master.NewID()
	if err != nil {
		return &meta.GateRegisterResponse{Header: &meta.ResponseHeader{Code: util.ErrorMasterNewID, Msg: err.Error()}}, nil
	}

	log.Debugf("Register user:%v password:%v", req.User, req.Password)

	if err = g.store.register(req.User, req.Password, id); err != nil {
		return &meta.GateRegisterResponse{Header: &meta.ResponseHeader{Code: util.ErrorRegister, Msg: err.Error()}}, nil
	}

	return &meta.GateRegisterResponse{ID: id}, nil
}

// UpdateUserInfo nickname, avatar.
func (g *Gate) UpdateUserInfo(ctx context.Context, req *meta.GateUpdateUserInfoRequest) (*meta.GateUpdateUserInfoResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	log.Debug("%d name:%v nickname:%v avatar:%v", c.getUser(), req.Name, req.NickName, req.Avatar)

	if req.NickName == "" && req.Avatar == "" {
		log.Errorf("%d name:%v nickname:%v avatar:%v", c.getUser(), req.Name, req.NickName, req.Avatar)
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserInfo, Msg: "nothing to update"}}, nil
	}

	if err := g.store.updateUserInfo(c.getUser(), req.Name, req.NickName, req.Avatar); err != nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserInfo, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateUserInfoResponse{}, nil
}

// UpdateSignature update user Signature
func (g *Gate) UpdateSignature(ctx context.Context, req *meta.GateUpdateSignatureRequest) (*meta.GateUpdateSignatureResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	if req.Signature == "" {
		log.Errorf("%d name:%v signature:%v", c.getUser(), req.Name, req.Signature)
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateSignature, Msg: "nothing to update"}}, nil
	}

	if err := g.store.updateSignature(c.getUser(), req.Name, req.Signature); err != nil {
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateSignature, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateSignatureResponse{}, nil
}

// UpdateUserPassword update user password
func (g *Gate) UpdateUserPassword(ctx context.Context, req *meta.GateUpdateUserPasswordRequest) (*meta.GateUpdateUserPasswordResponse, error) {
	c := g.manager.getConnection(ctx)
	if c != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	log.Debug("%d name:%v passwd old:%v new:%v", c.getUser(), req.Name, req.Password, req.NewPassword)
	if req.NewPassword == "" {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserPasswd, Msg: "new password is nil"}}, nil
	}

	if err := g.store.updateUserPassword(c.getUser(), req.Name, req.Password, req.NewPassword); err != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserPasswd, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateUserPasswordResponse{}, nil
}

// GetUserInfo get user base info
func (g *Gate) GetUserInfo(ctx context.Context, req *meta.GateGetUserInfoRequest) (*meta.GateGetUserInfoResponse, error) {
	c := g.manager.getConnection(ctx)
	if c != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	log.Debugf("%d get UserInfo byName:%v userName:%v userID:%v", c.getUser(), req.FindByName, req.UserName, req.UserID)

	userInfo, err := g.store.getUserInfo(c.getUser(), req.FindByName, req.UserName, req.UserID)
	if err != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetUserInfo, Msg: err.Error()}}, nil
	}
	log.Debugf("%d get UserInfo ByName:%v userName:%v userID:%v, name:%v, nickname:%v", c.getUser(), req.FindByName, req.UserName, req.UserID, userInfo.Name, userInfo.NickName)

	return &meta.GateGetUserInfoResponse{
		ID:        userInfo.ID,
		User:      userInfo.Name,
		NickName:  userInfo.NickName,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
	}, nil
}

// Login user,passwd.
func (g *Gate) Login(ctx context.Context, req *meta.GateUserLoginRequest) (*meta.GateUserLoginResponse, error) {
	log.Debugf("begin Login user:%v password:%v", req.User, req.Password)
	id, err := g.store.auth(req.User, req.Password)
	if err != nil {
		log.Errorf("end login error:%v", err)
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorAuth, Msg: err.Error()}}, nil
	}

	token, err := g.master.NewID()
	if err != nil {
		log.Errorf("end new id error:%v", err)
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorFailure, Msg: err.Error()}}, nil
	}

	g.manager.online(id, req.Device, token)

	return &meta.GateUserLoginResponse{ID: id, Token: token}, nil
}

// Logout nil.
func (g *Gate) Logout(ctx context.Context, req *meta.GateUserLogoutRequest) (*meta.GateUserLogoutResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateUserLogoutResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	g.manager.offline(c)

	return &meta.GateUserLogoutResponse{}, nil
}

// SendMessage 发送消息
func (g *Gate) SendMessage(ctx context.Context, req *meta.GateSendMessageRequest) (*meta.GateSendMessageResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: "login first"}}, nil
	}

	id, err := g.master.NewID()
	if err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	req.Msg.ID = id
	//防止乱写
	req.Msg.From = c.user

	if err = g.store.newMessage(c.user, *req.Msg); err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorNewMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateSendMessageResponse{ID: req.Msg.ID}, nil
}

// Stream 客户端登录后调用这个接口，接收下推消息
func (g *Gate) Stream(req *meta.GateStreamRequest, stream meta.Gate_StreamServer) error {
	c := g.manager.getConnByToken(req.Token)
	if c == nil {
		log.Errorf("not found connection, token:%d", req.Token)
		return ErrInvalidState
	}

	//如果这个函数返回了，说明连接断开了
	c.waitClose(stream)

	//如果有session就是登录过了，要清理资源
	g.manager.offline(c)

	log.Debugf("user:%d, token:%d timeout, offline", c.getUser(), c.getToken())

	return ErrInvalidState
}

// Friend 添加好友或确认接受添加.
func (g *Gate) Friend(ctx context.Context, req *meta.GateFriendRequest) (*meta.GateFriendResponse, error) {
	log.Debugf("begin Friend req:%v", req)
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	//自己要把自己添加成好友
	if req.UserID == c.getUser() {
		log.Infof("%d add friend id:%d", c.getUser(), req.UserID)
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorFriendSelf, Msg: "Friend ID must not be Self ID"}}, nil
	}

	if err := g.store.friend(c.getUser(), req.UserID, req.Operate, req.Msg); err != nil {
		log.Errorf("%d friend:%d operate:%d erorr:%s", c.getUser(), req.UserID, req.Operate, errors.ErrorStack(err))
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorAddFriend, Msg: err.Error()}}, nil
	}

	log.Debugf("end Friend req:%v", req)
	return &meta.GateFriendResponse{}, nil
}

// LoadFriendList 加载好友列表
func (g *Gate) LoadFriendList(ctx context.Context, req *meta.GateLoadFriendListRequest) (*meta.GateLoadFriendListResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	log.Debugf("%d begin loadFriendList", c.getUser())
	ids, err := g.store.loadFriendList(c.getUser())
	if err != nil {
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadFriendList, Msg: err.Error()}}, nil
	}
	log.Debugf("%d end loadFriendList:%v", c.getUser(), ids)

	return &meta.GateLoadFriendListResponse{Users: ids}, nil
}

// FindUser 添加好友前先查找,模糊查找
func (g *Gate) FindUser(ctx context.Context, req *meta.GateFindUserRequest) (*meta.GateFindUserResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}
	users, err := g.store.findUser(c.getUser(), req.User)
	if err != nil {
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorFindUser, Msg: err.Error()}}, nil
	}
	return &meta.GateFindUserResponse{Users: users}, nil
}

// GroupCreate 用户创建一个聊天组.
func (g *Gate) GroupCreate(ctx context.Context, req *meta.GateGroupCreateRequest) (*meta.GateGroupCreateResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	gid, err := g.master.NewID()
	if err != nil {
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorMasterNewID, Msg: err.Error()}}, nil
	}

	if err = g.store.groupCreate(c.getUser(), gid, req.Name); err != nil {
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("user:%d, create group:%d", c.getUser(), gid)
	return &meta.GateGroupCreateResponse{ID: gid}, nil
}

// GroupDelete 解散一个群.
func (g *Gate) GroupDelete(ctx context.Context, req *meta.GateGroupDeleteRequest) (*meta.GateGroupDeleteResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateGroupDeleteResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	if err := g.store.groupDelete(c.getUser(), req.ID); err != nil {
		return &meta.GateGroupDeleteResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("user:%d, delete group:%d", c.getUser(), req.ID)
	return &meta.GateGroupDeleteResponse{}, nil
}

// Group 添加，邀请，退出, 踢出
func (g *Gate) Group(ctx context.Context, req *meta.GateGroupRequest) (*meta.GateGroupResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	if err := g.store.group(c.getUser(), req.ID, req.Operate, req.Users, req.Msg); err != nil {
		return &meta.GateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("%d group:%d operate:%v, users:%v, msg:%v", c.getUser(), req.ID, req.Operate, req.Users, req.Msg)
	return &meta.GateGroupResponse{}, nil
}

// LoadGroupList 加载群组列表
func (g *Gate) LoadGroupList(ctx context.Context, req *meta.GateLoadGroupListRequest) (*meta.GateLoadGroupListResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	groups, err := g.store.loadGroupList(c.user)
	if err != nil {
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadGroup, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadGroupListResponse{Groups: groups}, nil
}

// UploadFile 客户端上传文件接口，一次一个文件.
func (g *Gate) UploadFile(ctx context.Context, req *meta.GateUploadFileRequest) (*meta.GateUploadFileResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	if err := g.store.uploadFile(c.user, req.File); err != nil {
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorUploadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateUploadFileResponse{}, nil
}

// CheckFile 客户端检测文件是否存在，文件的临时ID和md5, 服务器返回不存在的文件ID.
func (g *Gate) CheckFile(ctx context.Context, req *meta.GateCheckFileRequest) (*meta.GateCheckFileResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	names, err := g.store.checkFile(c.user, req.Names)
	if err != nil {
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorCheckFile, Msg: err.Error()}}, nil
	}

	return &meta.GateCheckFileResponse{Names: names}, nil
}

// DownloadFile 客户端下载文件，传入ID，返回具体文件内容.
func (g *Gate) DownloadFile(ctx context.Context, req *meta.GateDownloadFileRequest) (*meta.GateDownloadFileResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}

	files, err := g.store.downloadFile(c.user, req.Names)
	if err != nil {
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorDownloadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateDownloadFileResponse{Files: files}, nil
}

// LoadMessage 客户端同步离线消息，每次可逆序(旧消息)或正序(新消息)接收100条
func (g *Gate) LoadMessage(ctx context.Context, req *meta.GateLoadMessageRequest) (*meta.GateLoadMessageResponse, error) {
	c := g.manager.getConnection(ctx)
	if c == nil {
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: "login first"}}, nil
	}
	msgs, err := g.store.loadMessage(c.user, req.ID, req.Reverse)
	if err != nil {
		log.Errorf("%d loadMessage error:%s", c.user, errors.ErrorStack(err))
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadMessageResponse{Msgs: msgs}, nil
}

// Push notifer 调用的接口, 如果用户不在，要返回错误.
func (g *Gate) Push(ctx context.Context, req *meta.PushRequest) (*meta.PushResponse, error) {
	token, err := util.ContextGet(ctx, "token")
	if err != nil {
		log.Errorf("refuse token error:%s", err.Error())
		return &meta.PushResponse{Header: &meta.ResponseHeader{Code: -1, Msg: "ip refuse"}}, nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}

	refuse := true
	for _, i := range interfaces {
		if i.HardwareAddr != nil && i.HardwareAddr.String() == token {
			refuse = false
		}
	}

	if refuse {
		log.Errorf("refuse token:%s", token)
		return &meta.PushResponse{Header: &meta.ResponseHeader{Code: -1, Msg: "ip refuse"}}, nil
	}

	g.manager.pushMessage(req)
	return &meta.PushResponse{}, nil
}
