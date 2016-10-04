package store

import (
	"math"
	"net"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

// Store save user, message.
type Store struct {
	host    string
	dbPath  string
	user    *userDB
	group   *groupDB
	message *messageDB
	postman *postman
	friend  *friendDB
	file    *fileDB
	notice  *util.Notice
	master  *util.Master
}

// NewStore new Store server.
func NewStore(host, dbPath string) *Store {
	return &Store{
		host:    host,
		dbPath:  dbPath,
		user:    newUserDB(dbPath),
		message: newMessageDB(dbPath),
		group:   newGroupDB(dbPath),
		file:    newFileDB(dbPath),
	}
}

// Start Store service.
func (s *Store) Start(notice, master string) error {
	log.Debug("Store Start...")
	serv := grpc.NewServer()
	meta.RegisterStoreServer(serv, s)

	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	s.notice, err = util.NewNotice(notice)
	if err != nil {
		return errors.Trace(err)
	}

	s.master, err = util.NewMaster(master)
	if err != nil {
		return errors.Trace(err)
	}

	s.friend = newFriendDB(s.user)

	s.postman = newPostman(s.user, s.friend, s.group, s.notice)

	if err = s.user.start(); err != nil {
		return err
	}

	if err = s.group.start(); err != nil {
		return err
	}

	if err = s.file.start(); err != nil {
		return err
	}

	if err = s.message.start(s.postman); err != nil {
		return err
	}

	return serv.Serve(lis)
}

// Register add user.
func (s *Store) Register(_ context.Context, req *meta.StoreRegisterRequest) (*meta.StoreRegisterResponse, error) {
	log.Debugf("Store Register, user:%v passwd:%v ID:%v", req.User, req.Password, req.ID)
	if err := s.user.register(req.User, req.Password, req.ID); err != nil {
		return &meta.StoreRegisterResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreRegisterResponse{}, nil
}

// UpdateUserInfo update user base info, ex: nickname, picurl and so on
func (s *Store) UpdateUserInfo(_ context.Context, req *meta.StoreUpdateUserInfoRequest) (*meta.StoreUpdateUserInfoResponse, error) {
	log.Debugf("Store UpdateInfo, user:%v niceName:%v", req.User, req.NickName)
	id, err := s.user.updateUserInfo(req.User, req.NickName, req.Avatar)
	if err != nil {
		return &meta.StoreUpdateUserInfoResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreUpdateUserInfoResponse{ID: id}, nil
}

// UpdateUserPassword update user password
func (s *Store) UpdateUserPassword(_ context.Context, req *meta.StoreUpdateUserPasswordRequest) (*meta.StoreUpdateUserPasswordResponse, error) {
	log.Debugf("Store UpdatePassword, user:")
	id, err := s.user.updateUserPassword(req.User, req.Password)
	if err != nil {
		return &meta.StoreUpdateUserPasswordResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreUpdateUserPasswordResponse{ID: id}, nil
}

// GetUserInfo get user base info
func (s *Store) GetUserInfo(_ context.Context, req *meta.StoreGetUserInfoRequest) (*meta.StoreGetUserInfoResponse, error) {
	log.Debugf("GetUserInfo, type:%v userName:%v userID:%v", req.Type, req.UserName, req.UserID)
	a, err := s.user.getUserInfo(req.Type, req.UserName, req.UserID)
	if err != nil {
		return &meta.StoreGetUserInfoResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreGetUserInfoResponse{ID: a.ID, User: a.Name, NickName: a.NickName, Avatar: a.Avatar}, nil
}

// Auth check password.
func (s *Store) Auth(_ context.Context, req *meta.StoreAuthRequest) (*meta.StoreAuthResponse, error) {
	log.Debugf("Store Auth, user:%v passwd:%v", req.User, req.Password)
	id, err := s.user.auth(req.User, req.Password)
	if err != nil {
		return &meta.StoreAuthResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreAuthResponse{ID: id}, nil
}

// FindUser 根据字符串的用户名模糊查询用户信息.
func (s *Store) FindUser(_ context.Context, req *meta.StoreFindUserRequest) (*meta.StoreFindUserResponse, error) {
	log.Debugf("Store FindUser, user:%v", req.User)
	users, err := s.user.findUser(req.User)
	if err != nil {
		return &meta.StoreFindUserResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.StoreFindUserResponse{Users: users}, nil
}

// Friend 添加好友添加完后会返回当前好友关系状态.
func (s *Store) Friend(_ context.Context, req *meta.StoreFriendRequest) (*meta.StoreFriendResponse, error) {
	log.Debugf("Store Friend from:%v to:%v Operate:%v", req.From, req.To, req.Operate)
	//这些事件都需要创建一个给对方的消息
	var err error
	switch req.Operate {
	case meta.Relation_ADD:
		//1.存储一个添加对方为好友的消息
		//2.要给对方一个请求添加好友的消息
		err = s.user.friend.set(req.From, req.To, req.Operate, req.Msg)

	case meta.Relation_CONFIRM:
		//1.存储一个确认添加为好友的消息
		//2.要给对方一个确认添加好友的消息
		err = s.user.friend.set(req.From, req.To, req.Operate, req.Msg)

	case meta.Relation_REFUSE:
		//只需要给对方一个拒绝的消息就行

	case meta.Relation_DEL:
		//只需要给对方发个通知
		er = s.user.friend.remove(req.From, req.To)
	}

	if err != nil {
		log.Debugf("Store Friend from:%v to:%v Operate:%v, error:%v", req.From, req.To, req.Operate, errors.ErrorStack(err))
		return &meta.StoreFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	// 发个提示消息
	id, err := s.master.NewID()
	if err != nil {
		log.Debugf("Store Friend from:%v to:%v Operate:%v, error:%v", req.From, req.To, req.Operate, errors.ErrorStack(err))
		return &meta.StoreFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	pm := meta.PushMessage{Event: meta.Event_Friend, Operate: req.Operate, Msg: &meta.Message{ID: id, From: req.From, To: req.To, Body: req.Msg}}
	// 直接发送，如果失败会自动插入到重试队列中
	if err := s.message.send(pm); err != nil {
		log.Debugf("Store Friend from:%v to:%v Operate:%v, error:%v", req.From, req.To, req.Operate, errors.ErrorStack(err))
		return &meta.StoreFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	return &meta.StoreFriendResponse{}, nil
}

// LoadFriendList load user's friend list
func (s *Store) LoadFriendList(_ context.Context, req *meta.StoreLoadFriendListRequest) (*meta.StoreLoadFriendListResponse, error) {
	ids, err := s.friend.get(req.User)
	if err != nil {
		return &meta.StoreLoadFriendListResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	return &meta.StoreLoadFriendListResponse{Users: ids}, nil
}

// NewMessage save message to leveldb,
func (s *Store) NewMessage(_ context.Context, req *meta.StoreNewMessageRequest) (*meta.StoreNewMessageResponse, error) {
	log.Debugf("Store NewMessage, msg:%v", req.Msg)
	// 直接发送，如果失败会自动插入到重试队列中
	if err := s.message.send(meta.PushMessage{Msg: req.Msg}); err != nil {
		return &meta.StoreNewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	log.Debugf("Store messate success")

	return &meta.StoreNewMessageResponse{}, nil
}

// UploadFile 上传文件接口，一次一个文件.
func (s *Store) UploadFile(_ context.Context, req *meta.StoreUploadFileRequest) (*meta.StoreUploadFileResponse, error) {
	key := util.MD5(req.File)
	if err := s.file.add(key, req.File); err != nil {
		return &meta.StoreUploadFileResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.StoreUploadFileResponse{}, nil
}

// CheckFile 检测文件是否存在，文件的MD5, 服务器返回不存在的文件MD5.
func (s *Store) CheckFile(_ context.Context, req *meta.StoreCheckFileRequest) (*meta.StoreCheckFileResponse, error) {
	var names []string
	for _, name := range req.Names {
		ok, err := s.file.exist([]byte(name))
		if err != nil {
			return &meta.StoreCheckFileResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}
		if !ok {
			names = append(names, name)
		}
	}
	return &meta.StoreCheckFileResponse{Names: names}, nil
}

// DownloadFile 下载文件，传入文件MD5，返回具体文件内容.
func (s *Store) DownloadFile(_ context.Context, req *meta.StoreDownloadFileRequest) (*meta.StoreDownloadFileResponse, error) {
	files := make(map[string][]byte)
	for _, name := range req.Names {
		data, err := s.file.get([]byte(name))
		if err != nil {
			return &meta.StoreDownloadFileResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}
		files[name] = data
	}

	return &meta.StoreDownloadFileResponse{Files: files}, nil
}

// LoadMessage 收取用户消息，每次可逆序(旧消息)或正序(新消息)接收100条
// 如果ID为0，就逆序返回旧的100条消息
func (s *Store) LoadMessage(_ context.Context, req *meta.StoreLoadMessageRequest) (*meta.StoreLoadMessageResponse, error) {
	// 修正下，确定前面不会传错
	if req.ID == 0 {
		req.ID = math.MaxInt64
		req.Reverse = true
	}

	ids, err := s.user.getMessage(req.User, req.Reverse, req.ID)
	if err != nil {
		return &meta.StoreLoadMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	msgs, err := s.message.get(ids...)
	if err != nil {
		return &meta.StoreLoadMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreLoadMessageResponse{Msgs: msgs}, nil
}

// CreateGroup 创建群组
func (s *Store) CreateGroup(_ context.Context, req *meta.StoreCreateGroupRequest) (*meta.StoreCreateGroupResponse, error) {
	//创建群组
	group := meta.GroupInfo{ID: req.GroupID, Name: req.GroupName}
	err := s.group.newGroup(group)
	if err != nil {
		return &meta.StoreCreateGroupResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	//TODO 还需要考虑群主和管理员的处理
	//向群组中添加用户
	err = s.group.addUser(group.ID, req.UserID)
	if err != nil {
		return &meta.StoreCreateGroupResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	//在用户数据库记录群组信息
	err = s.user.addGroup(req.UserID, req.GroupID)
	if err != nil {
		return &meta.StoreCreateGroupResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.StoreCreateGroupResponse{}, nil
}

// LoadGroupList 加载群组列表
func (s *Store) LoadGroupList(_ context.Context, req *meta.StoreLoadGroupListRequest) (*meta.StoreLoadGroupListResponse, error) {

	//获取群组id列表
	gids, err := s.user.getGroups(req.User)
	if err != nil {
		return &meta.StoreLoadGroupListResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	var groups []*meta.GroupInfo
	//分别获取群组信息
	for _, gid := range gids {
		g, err := s.group.get(gid)
		if err != nil {
			return &meta.StoreLoadGroupListResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}

		groups = append(groups, &g)
	}

	return &meta.StoreLoadGroupListResponse{Groups: groups}, nil
}
