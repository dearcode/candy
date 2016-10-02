package gate

import (
	"net"
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

var (
	// ErrUndefineMethod 方法未定义.
	ErrUndefineMethod = errors.New("undefine method")
	// ErrInvalidContext 从context中解析客户端地址时出错.
	ErrInvalidContext = errors.New("invalid context")
	// ErrInvalidState 当前用户离线或未登录.
	ErrInvalidState = errors.New("invalid context")
)

// Gate recv client request.
type Gate struct {
	host     string
	store    *store
	master   *util.Master
	notice   *util.Notice
	sessions map[string]*session
	ids      map[int64]*session
	sync.RWMutex
	healthServer *health.Server // nil means disabled
}

// NewGate new gate server.
func NewGate() *Gate {
	return &Gate{
		sessions:     make(map[string]*session),
		ids:          make(map[int64]*session),
		healthServer: health.NewServer(),
	}
}

// Start Gate service.
func (g *Gate) Start(host, notice, master, store string) error {
	log.Debugf("Gate Start...")

	g.host = host

	lis, err := net.Listen("tcp", host)
	if err != nil {
		return errors.Trace(err)
	}

	g.notice, err = util.NewNotice(notice)
	if err != nil {
		return errors.Trace(err)
	}

	g.master, err = util.NewMaster(master)
	if err != nil {
		return errors.Trace(err)
	}

	g.store = newStore(store)
	if err = g.store.start(); err != nil {
		return errors.Trace(err)
	}

	serv := grpc.NewServer()
	meta.RegisterGateServer(serv, g)

	if g.healthServer != nil {
		healthpb.RegisterHealthServer(serv, g.healthServer)
	}

	return serv.Serve(lis)
}

func (g *Gate) getSession(ctx context.Context) (*session, error) {
	log.Debug("Gate getSession")
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.Trace(ErrInvalidContext)
	}

	addrs, ok := md["remote"]
	if !ok {
		return nil, errors.Trace(ErrInvalidContext)
	}

	g.RLock()
	s, ok := g.sessions[addrs[0]]
	g.RUnlock()

	if !ok {
		s = newSession(addrs[0])
		g.Lock()
		g.sessions[addrs[0]] = s
		g.Unlock()
	}

	return s, nil
}

func (g *Gate) online(s *session, id int64) {
	s.online(id)
	g.Lock()
	g.ids[s.id] = s
	g.Unlock()
}

func (g *Gate) getOnlineUser(id int64) *session {
	g.RLock()
	s, ok := g.ids[id]
	g.RUnlock()

	if ok {
		return s
	}

	return nil
}

func (g *Gate) offline(s *session) {
	s.offline()
	g.Lock()
	delete(g.ids, s.id)
	g.Unlock()
}

func (g *Gate) getOnlineSession(ctx context.Context) (*session, error) {
	log.Debug("Gate getOnlineSession")
	s, err := g.getSession(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !s.isOnline() {
		return nil, ErrInvalidState
	}

	return s, nil
}

// Register user, passwd.
func (g *Gate) Register(ctx context.Context, req *meta.GateRegisterRequest) (*meta.GateRegisterResponse, error) {
	log.Debug("Gate Register")
	_, err := g.getSession(ctx)
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

// UpdateUserInfo nickname.
func (g *Gate) UpdateUserInfo(ctx context.Context, req *meta.GateUpdateUserInfoRequest) (*meta.GateUpdateUserInfoResponse, error) {
	log.Debug("Gate UpdateUserInfo")
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if !s.isOnline() {
		err := errors.Errorf("current user is offline")
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorOffline, Msg: err.Error()}}, nil
	}

	log.Debugf("updateUserInfo user:%v niceName:%v", req.User, req.NickName)
	id, err := g.store.updateUserInfo(req.User, req.NickName, req.Avatar)
	if err != nil {
		return &meta.GateUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserInfo, Msg: err.Error()}, ID: id}, nil
	}

	return &meta.GateUpdateUserInfoResponse{ID: id}, nil
}

// UpdateUserPassword update user password
func (g *Gate) UpdateUserPassword(ctx context.Context, req *meta.GateUpdateUserPasswordRequest) (*meta.GateUpdateUserPasswordResponse, error) {
	log.Debug("Gate UpdateUserPassword")
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if !s.isOnline() {
		err := errors.Errorf("current user is offline")
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorOffline, Msg: err.Error()}}, nil
	}

	log.Debugf("updateUserPassword user:%v", req.User)
	id, err := g.store.updateUserPassword(req.User, req.Password)
	if err != nil {
		return &meta.GateUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: util.ErrorUpdateUserPasswd, Msg: err.Error()}}, nil
	}

	return &meta.GateUpdateUserPasswordResponse{ID: id}, nil
}

// GetUserInfo get user base info
func (g *Gate) GetUserInfo(ctx context.Context, req *meta.GateGetUserInfoRequest) (*meta.GateGetUserInfoResponse, error) {
	log.Debugf("Gate UserInfo")
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if !s.isOnline() {
		err := errors.Errorf("current user is offline")
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorOffline, Msg: err.Error()}}, nil
	}

	log.Debugf("get UserInfo type:%v userName:%v userID:%v", req.Type, req.UserName, req.UserID)
	id, name, nickName, avatar, err := g.store.getUserInfo(req.Type, req.UserName, req.UserID)
	if err != nil {
		return &meta.GateGetUserInfoResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetUserInfo, Msg: err.Error()}}, nil
	}

	return &meta.GateGetUserInfoResponse{ID: id, User: name, NickName: nickName, Avatar: avatar}, nil
}

// Login user,passwd.
func (g *Gate) Login(ctx context.Context, req *meta.GateUserLoginRequest) (*meta.GateUserLoginResponse, error) {
	log.Debug("Gate Login")
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	log.Debugf("Login user:%v password:%v", req.User, req.Password)
	id, err := g.store.auth(req.User, req.Password)
	if err != nil {
		log.Errorf("auth error:%v", err)
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorAuth, Msg: err.Error()}}, nil
	}

	g.online(s, id)

	//订阅消息
	if err := g.notice.Subscribe(s.getID(), g.host); err != nil {
		return &meta.GateUserLoginResponse{Header: &meta.ResponseHeader{Code: util.ErrorSubscribe, Msg: err.Error()}}, nil
	}

	return &meta.GateUserLoginResponse{ID: id}, nil
}

// Logout nil.
func (g *Gate) Logout(ctx context.Context, req *meta.GateUserLogoutRequest) (*meta.GateUserLogoutResponse, error) {
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateUserLogoutResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	//注销需要先取消消息订阅
	err = g.notice.UnSubscribe(s.getID(), g.host)
	if err != nil {
		return &meta.GateUserLogoutResponse{Header: &meta.ResponseHeader{Code: util.ErrorUnSubscribe, Msg: err.Error()}}, nil
	}

	g.offline(s)
	return &meta.GateUserLogoutResponse{}, nil
}

// SendMessage 发送消息
func (g *Gate) SendMessage(ctx context.Context, req *meta.GateSendMessageRequest) (*meta.GateSendMessageResponse, error) {
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetSession, Msg: err.Error()}}, nil
	}

	if req.Msg.ID, err = g.master.NewID(); err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	//防止乱写
	req.Msg.From = s.getID()

	if err = g.store.newMessage(req.Msg); err != nil {
		return &meta.GateSendMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorNewMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateSendMessageResponse{Id: req.Msg.ID}, nil
}

// Ready 连接成功后立刻调用Ready, 开启推送
func (g *Gate) Ready(msg *meta.Message, stream meta.Gate_ReadyServer) error {
	s, err := g.getSession(stream.Context())
	if err != nil {
		return errors.Trace(err)
	}

	for {
		m := <-s.push
		log.Debugf("recv push request m:%v", m)
		stream.Send(m)
	}
}

// Heartbeat nil.
func (g *Gate) Heartbeat(ctx context.Context, req *meta.GateHeartbeatRequest) (*meta.GateHeartbeatResponse, error) {
	s, err := g.getSession(ctx)
	if err != nil {
		return &meta.GateHeartbeatResponse{}, nil
	}

	//已经离线就不在处理
	if !s.isOnline() {
		return &meta.GateHeartbeatResponse{}, nil
	}

	//更新心跳信息
	s.heartbeat()

	return &meta.GateHeartbeatResponse{}, nil
}

// Push recv Notice server Message, and send Message to client.
func (g *Gate) Push(ctx context.Context, req *meta.GatePushRequest) (*meta.GatePushResponse, error) {
	log.Debugf("begin PushID:%v msg:%v", req.ID, req.Msg)

	for _, id := range req.ID {
		s := g.getOnlineUser(id.User)
		if s == nil {
			log.Debugf("User:%d offline", id.User)
			continue
		}
		req.Msg.Msg.Before = id.Before
		log.Debugf("will push user:%d, msg:%v", id.User, req.Msg)
		s.push <- req.Msg
	}

	log.Debugf("end PushID:%v msg:%v ok", req.ID, req.Msg)

	return &meta.GatePushResponse{}, nil
}

// Friend 添加好友或确认接受添加.
func (g *Gate) Friend(ctx context.Context, req *meta.GateFriendRequest) (*meta.GateFriendResponse, error) {
	log.Debugf("begin Friend req:%v", req)
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	//自己要把自己添加成好友
	if req.UserID == s.id {
		log.Infof("%d add friend id:%d", s.id, req.UserID)
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorFriendSelf, Msg: "Friend ID must not be Self ID"}}, nil
	}

	if err = g.store.friend(s.id, req.UserID, req.Operate, req.Msg); err != nil {
		log.Errorf("%d friend:%d operate:%d erorr:%s", s.id, req.UserID, req.Operate, errors.ErrorStack(err))
		return &meta.GateFriendResponse{Header: &meta.ResponseHeader{Code: util.ErrorAddFriend, Msg: err.Error()}}, nil
	}

	log.Debugf("end Friend req:%v", req)
	return &meta.GateFriendResponse{}, nil
}

// LoadFriendList 加载好友列表
func (g *Gate) LoadFriendList(ctx context.Context, req *meta.GateLoadFriendListRequest) (*meta.GateLoadFriendListResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	ids, err := g.store.loadFriendList(s.id)
	if err != nil {
		return &meta.GateLoadFriendListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadFriendList, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadFriendListResponse{Users: ids}, nil
}

// FindUser 添加好友前先查找,模糊查找
func (g *Gate) FindUser(ctx context.Context, req *meta.GateFindUserRequest) (*meta.GateFindUserResponse, error) {
	_, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}
	users, err := g.store.findUser(req.User)
	if err != nil {
		return &meta.GateFindUserResponse{Header: &meta.ResponseHeader{Code: util.ErrorFindUser, Msg: err.Error()}}, nil
	}
	return &meta.GateFindUserResponse{Users: users}, nil
}

// CreateGroup 用户创建一个聊天组.
func (g *Gate) CreateGroup(ctx context.Context, req *meta.GateCreateGroupRequest) (*meta.GateCreateGroupResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateCreateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	gid, err := g.master.NewID()
	if err != nil {
		return &meta.GateCreateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorMasterNewID, Msg: err.Error()}}, nil
	}

	if err = g.store.createGroup(s.getID(), gid, req.GroupName); err != nil {
		return &meta.GateCreateGroupResponse{Header: &meta.ResponseHeader{Code: util.ErrorCreateGroup, Msg: err.Error()}}, nil
	}

	log.Debugf("user:%d, create group:%d", s.getID(), gid)
	return &meta.GateCreateGroupResponse{ID: gid}, nil
}

// LoadGroupList 加载群组列表
func (g *Gate) LoadGroupList(ctx context.Context, req *meta.GateLoadGroupListRequest) (*meta.GateLoadGroupListResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	groups, err := g.store.loadGroupList(s.id)
	if err != nil {
		return &meta.GateLoadGroupListResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadGroup, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadGroupListResponse{Groups: groups}, nil
}

// UploadFile 客户端上传文件接口，一次一个文件.
func (g *Gate) UploadFile(ctx context.Context, req *meta.GateUploadFileRequest) (*meta.GateUploadFileResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	if err = g.store.uploadFile(s.id, req.File); err != nil {
		return &meta.GateUploadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorUploadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateUploadFileResponse{}, nil
}

// CheckFile 客户端检测文件是否存在，文件的临时ID和md5, 服务器返回不存在的文件ID.
func (g *Gate) CheckFile(ctx context.Context, req *meta.GateCheckFileRequest) (*meta.GateCheckFileResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	names, err := g.store.checkFile(s.id, req.Names)
	if err != nil {
		return &meta.GateCheckFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorCheckFile, Msg: err.Error()}}, nil
	}

	return &meta.GateCheckFileResponse{Names: names}, nil
}

// DownloadFile 客户端下载文件，传入ID，返回具体文件内容.
func (g *Gate) DownloadFile(ctx context.Context, req *meta.GateDownloadFileRequest) (*meta.GateDownloadFileResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}

	files, err := g.store.downloadFile(s.id, req.Names)
	if err != nil {
		return &meta.GateDownloadFileResponse{Header: &meta.ResponseHeader{Code: util.ErrorDownloadFile, Msg: err.Error()}}, nil
	}

	return &meta.GateDownloadFileResponse{Files: files}, nil
}

// LoadMessage 客户端同步离线消息，每次可逆序(旧消息)或正序(新消息)接收100条
func (g *Gate) LoadMessage(ctx context.Context, req *meta.GateLoadMessageRequest) (*meta.GateLoadMessageResponse, error) {
	s, err := g.getOnlineSession(ctx)
	if err != nil {
		log.Errorf("getSession error:%s", errors.ErrorStack(err))
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorGetOnlineSession, Msg: err.Error()}}, nil
	}
	msgs, err := g.store.loadMessage(s.id, req.ID, req.Reverse)
	if err != nil {
		log.Errorf("%d loadMessage error:%s", s.id, errors.ErrorStack(err))
		return &meta.GateLoadMessageResponse{Header: &meta.ResponseHeader{Code: util.ErrorLoadMessage, Msg: err.Error()}}, nil
	}

	return &meta.GateLoadMessageResponse{Msgs: msgs}, nil
}
