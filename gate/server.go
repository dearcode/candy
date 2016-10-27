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
		manager:      newManager(nc),
		store:        sc,
		master:       mc,
		healthServer: health.NewServer(),
		server:       grpc.NewServer(),
	}
	meta.RegisterGateServer(g.server, g)

	healthpb.RegisterHealthServer(g.server, g.healthServer)

	return g, g.server.Serve(l)
}

// Register user, passwd.
func (g *Gate) Register(ctx context.Context, req *meta.GateRegisterRequest) (*meta.GateRegisterResponse, error) {
	log.Debug("Gate Register")
	_, _, err := g.manager.getConnection(ctx)
	if err != nil {
		return &meta.GateRegisterResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

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
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}
	log.Debug("%d name:%v nickname:%v avatar:%v", s.user, req.Name, req.NickName, req.Avatar)

	if req.NickName == "" && req.Avatar == "" {
		log.Errorf("%d name:%v nickname:%v avatar:%v", s.user, req.Name, req.NickName, req.Avatar)
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserInfo, Msg: err.Error()}}, nil
	}

	if err = g.store.updateUserInfo(s.user, req.Name, req.NickName, req.Avatar); err != nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserInfo, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateUserInfoResponse{}, nil
}

// UpdateSignature update user Signature
func (g *Gate) UpdateSignature(ctx context.Context, req *meta.GateUpdateSignatureRequest) (*meta.GateUpdateSignatureResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if req.Signature == "" {
		log.Errorf("%d name:%v signature:%v", s.user, req.Name, req.Signature)
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateSignature, Msg: err.Error()}}, nil
	}

	if err = g.store.updateSignature(s.user, req.Name, req.Signature); err != nil {
		return &meta.GateUpdateSignatureResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateSignature, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateSignatureResponse{}, nil
}

// UpdateUserPassword update user password
func (g *Gate) UpdateUserPassword(ctx context.Context, req *meta.GateUpdateUserPasswordRequest) (*meta.GateUpdateUserPasswordResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	log.Debug("%d name:%v passwd old:%v new:%v", s.user, req.Name, req.Password, req.NewPassword)
	if req.NewPassword == "" {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserPasswd, Msg: "new password is nil"}}, nil
	}

	if err = g.store.updateUserPassword(s.user, req.Name, req.Password, req.NewPassword); err != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserPasswd, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateUserPasswordResponse{}, nil
}

// GetUserInfo get user base info
func (g *Gate) GetUserInfo(ctx context.Context, req *meta.GateGetUserInfoRequest) (*meta.GateGetUserInfoResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	log.Debugf("%d get UserInfo byName:%v userName:%v userID:%v", s.user, req.FindByName, req.UserName, req.UserID)

	userInfo, err := g.store.getUserInfo(s.user, req.FindByName, req.UserName, req.UserID)
	if err != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetUserInfo, Msg: err.Error()}}, nil
	}
	log.Debugf("%d get UserInfo ByName:%v userName:%v userID:%v, name:%v, nickname:%v", s.user, req.FindByName, req.UserName, req.UserID, userInfo.Name, userInfo.NickName)

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
	log.Debug("Gate Login")
	c, _, err := g.manager.getConnection(ctx)
	if err != nil {
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	log.Debugf("Login user:%v password:%v", req.User, req.Password)
	id, err := g.store.auth(req.User, req.Password)
	if err != nil {
		log.Errorf("auth error:%v", err)
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorAuth, Msg: err.Error()}}, nil
	}

	g.manager.online(id, req.Device, c)

	return &meta.GateUserLoginResponse{ID: id}, nil
}

// Logout nil.
func (g *Gate) Logout(ctx context.Context, req *meta.GateUserLogoutRequest) (*meta.GateUserLogoutResponse, error) {
	s, c, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateUserLogoutResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	g.manager.offline(s.user, c)

	return &meta.GateUserLogoutResponse{}, nil
}

// SendMessage 发送消息
func (g *Gate) SendMessage(ctx context.Context, req *meta.GateSendMessageRequest) (*meta.GateSendMessageResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if req.Msg.ID, err = g.master.NewID(); err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	//防止乱写
	req.Msg.From = s.user

	if err = g.store.newMessage(s.user, *req.Msg); err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorNewMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateSendMessageResponse{ID: req.Msg.ID}, nil
}

// Stream 连接成功后立刻调用Stream, 开启推送
func (g *Gate) Stream(msg *meta.Message, stream meta.Gate_StreamServer) error {
	c, _, err := g.manager.getConnection(stream.Context())
	if err != nil {
		return errors.Trace(err)
	}

	c.waitClose(stream)

	return nil
}

// Heartbeat nil.
func (g *Gate) Heartbeat(ctx context.Context, req *meta.GateHeartbeatRequest) (*meta.GateHeartbeatResponse, error) {
	_, c, err := g.manager.getSession(ctx)
	if err != nil {
		return &meta.GateHeartbeatResponse{}, nil
	}

	//更新心跳信息
	c.heartbeat()

	return &meta.GateHeartbeatResponse{}, nil
}

// Friend 添加好友或确认接受添加.
func (g *Gate) Friend(ctx context.Context, req *meta.GateFriendRequest) (*meta.GateFriendResponse, error) {
	log.Debugf("begin Friend req:%v", req)
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Infof("get session error:%s", errors.ErrorStack(err))
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	//自己要把自己添加成好友
	if req.UserID == s.user {
		log.Infof("%d add friend id:%d", s.user, req.UserID)
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorFriendSelf, Msg: "Friend ID must not be Self ID"}}, nil
	}

	if err = g.store.friend(s.user, req.UserID, req.Operate, req.Msg); err != nil {
		log.Errorf("%d friend:%d operate:%d erorr:%s", s.user, req.UserID, req.Operate, errors.ErrorStack(err))
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorAddFriend, Msg: err.Error()}}, nil
	}

	log.Debugf("end Friend req:%v", req)
	return &meta.GateFriendResponse{}, nil
}

// LoadFriendList 加载好友列表
func (g *Gate) LoadFriendList(ctx context.Context, req *meta.GateLoadFriendListRequest) (*meta.GateLoadFriendListResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	log.Debugf("%d begin loadFriendList", s.user)
	ids, err := g.store.loadFriendList(s.user)
	if err != nil {
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadFriendList, Msg: err.Error()}}, nil
	}
	log.Debugf("%d end loadFriendList:%v", s.user, ids)

	return &meta.GateLoadFriendListResponse{Users: ids}, nil
}

// FindUser 添加好友前先查找,模糊查找
func (g *Gate) FindUser(ctx context.Context, req *meta.GateFindUserRequest) (*meta.GateFindUserResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}
	users, err := g.store.findUser(s.user, req.User)
	if err != nil {
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorFindUser, Msg: err.Error()}}, nil
	}
	return &meta.GateFindUserResponse{Users: users}, nil
}

// GroupCreate 用户创建一个聊天组.
func (g *Gate) GroupCreate(ctx context.Context, req *meta.GateGroupCreateRequest) (*meta.GateGroupCreateResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	gid, err := g.master.NewID()
	if err != nil {
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorMasterNewID, Msg: err.Error()}}, nil
	}

	if err = g.store.groupCreate(s.user, gid, req.Name); err != nil {
		return &meta.GateGroupCreateResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("user:%d, create group:%d", s.user, gid)
	return &meta.GateGroupCreateResponse{ID: gid}, nil
}

// GroupDelete 解散一个群.
func (g *Gate) GroupDelete(ctx context.Context, req *meta.GateGroupDeleteRequest) (*meta.GateGroupDeleteResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateGroupDeleteResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	if err = g.store.groupDelete(s.user, req.ID); err != nil {
		return &meta.GateGroupDeleteResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("user:%d, delete group:%d", s.user, req.ID)
	return &meta.GateGroupDeleteResponse{}, nil
}

// Group 添加，邀请，退出, 踢出
func (g *Gate) Group(ctx context.Context, req *meta.GateGroupRequest) (*meta.GateGroupResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	if err = g.store.group(s.user, req.ID, req.Operate, req.Users, req.Msg); err != nil {
		return &meta.GateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("%d group:%d operate:%v, users:%v, msg:%v", s.user, req.ID, req.Operate, req.Users, req.Msg)
	return &meta.GateGroupResponse{}, nil
}

// LoadGroupList 加载群组列表
func (g *Gate) LoadGroupList(ctx context.Context, req *meta.GateLoadGroupListRequest) (*meta.GateLoadGroupListResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	groups, err := g.store.loadGroupList(s.user)
	if err != nil {
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadGroup, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadGroupListResponse{Groups: groups}, nil
}

// UploadFile 客户端上传文件接口，一次一个文件.
func (g *Gate) UploadFile(ctx context.Context, req *meta.GateUploadFileRequest) (*meta.GateUploadFileResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	if err = g.store.uploadFile(s.user, req.File); err != nil {
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorUploadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateUploadFileResponse{}, nil
}

// CheckFile 客户端检测文件是否存在，文件的临时ID和md5, 服务器返回不存在的文件ID.
func (g *Gate) CheckFile(ctx context.Context, req *meta.GateCheckFileRequest) (*meta.GateCheckFileResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	names, err := g.store.checkFile(s.user, req.Names)
	if err != nil {
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorCheckFile, Msg: err.Error()}}, nil
	}

	return &meta.GateCheckFileResponse{Names: names}, nil
}

// DownloadFile 客户端下载文件，传入ID，返回具体文件内容.
func (g *Gate) DownloadFile(ctx context.Context, req *meta.GateDownloadFileRequest) (*meta.GateDownloadFileResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	files, err := g.store.downloadFile(s.user, req.Names)
	if err != nil {
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorDownloadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateDownloadFileResponse{Files: files}, nil
}

// LoadMessage 客户端同步离线消息，每次可逆序(旧消息)或正序(新消息)接收100条
func (g *Gate) LoadMessage(ctx context.Context, req *meta.GateLoadMessageRequest) (*meta.GateLoadMessageResponse, error) {
	s, _, err := g.manager.getSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}
	msgs, err := g.store.loadMessage(s.user, req.ID, req.Reverse)
	if err != nil {
		log.Errorf("%d loadMessage error:%s", s.user, errors.ErrorStack(err))
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadMessageResponse{Msgs: msgs}, nil
}
