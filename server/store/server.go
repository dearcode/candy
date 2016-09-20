package store

import (
	"net"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
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
	log.Debugf("GetUserInfo, user:%v", req.User)
	a, err := s.user.getUserInfo(req.User)
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

// AddFriend 添加好友添加完后会返回当前好友关系状态.
func (s *Store) AddFriend(_ context.Context, req *meta.StoreAddFriendRequest) (*meta.StoreAddFriendResponse, error) {
	log.Debugf("Store AddFriend, from:%v to:%v State:%v", req.From, req.To, req.State)
	state, err := s.user.friend.add(req.From, req.To, req.State)
	if err != nil {
		return &meta.StoreAddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	// 主动的添加消息就不需要推送了
	if req.State == meta.FriendRelation_Active {
		return &meta.StoreAddFriendResponse{State: state}, nil
	}

	// 发个提示消息
	id, err := s.master.NewID()
	if err != nil {
		return &meta.StoreAddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	msg := meta.Message{ID: id, Method: meta.Method_FRIEND_ADD, From: req.To, Body: req.Msg}
	if state == meta.FriendRelation_Confirm {
		msg.Method = meta.Method_FRIEND_CONFIRM
	}

	// 发通知
	if err = s.notice.Push(msg, req.From); err != nil {
		return &meta.StoreAddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: errors.ErrorStack(err)}}, nil
	}

	return &meta.StoreAddFriendResponse{State: state}, nil
}

// CreateGroup create group.
func (s *Store) CreateGroup(_ context.Context, req *meta.StoreCreateGroupRequest) (*meta.StoreCreateGroupResponse, error) {
	log.Debugf("Store CreateGroup")
	return nil, nil
}

// NewMessage save message to leveldb,
func (s *Store) NewMessage(_ context.Context, req *meta.StoreNewMessageRequest) (*meta.StoreNewMessageResponse, error) {
	log.Debugf("Store NewMessage, msg:%v", req.Msg)
	// add消息到db
	if err := s.message.add(*req.Msg); err != nil {
		return &meta.StoreNewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	log.Debugf("add message to db success")
	// 再添加未推送消息队列
	if err := s.message.addQueue(req.Msg.ID); err != nil {
		return &meta.StoreNewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	log.Debugf("add messate to queue success")
	// 再调用推送
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
