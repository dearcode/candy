package store

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util/log"
)

// Store save user, message.
type Store struct {
	host    string
	notice  string
	dbPath  string
	user    *userDB
	group   *groupDB
	message *messageDB
	postman *postman
	friend  *friendDB
}

// NewStore new Store server.
func NewStore(host, notice, dbPath string) *Store {
	s := &Store{
		host:    host,
		dbPath:  dbPath,
		user:    newUserDB(dbPath),
		message: newMessageDB(dbPath),
		group:   newGroupDB(dbPath),
	}

	s.friend = newFriendDB(s.user)
	s.postman = newPostman(notice, s.user, s.friend, s.group)

	return s
}

// Start Store service.
func (s *Store) Start() error {
	log.Debug("Store Start...")
	serv := grpc.NewServer()
	meta.RegisterStoreServer(serv, s)

	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	if err = s.user.start(); err != nil {
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

// FindUser 根据字符串的用户名查的用户信息.
func (s *Store) FindUser(_ context.Context, req *meta.StoreFindUserRequest) (*meta.StoreFindUserResponse, error) {
	log.Debugf("Store FindUser, user:%v", req.User)
	id, err := s.user.findUser(req.User)
	if err != nil {
		return &meta.StoreFindUserResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.StoreFindUserResponse{ID: id}, nil
}

// AddFriend 添加好友，两人都添加过对方后才可以聊天.
func (s *Store) AddFriend(_ context.Context, req *meta.StoreAddFriendRequest) (*meta.StoreAddFriendResponse, error) {
	log.Debugf("Store AddFriend, from:%v to:%v Confirm:%v", req.From, req.To, req.Confirm)
	if err := s.user.friend.add(req.From, req.To, req.Confirm); err != nil {
		return &meta.StoreAddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	if req.Confirm {
		if err := s.user.friend.add(req.To, req.From, req.Confirm); err != nil {
			return &meta.StoreAddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}
		return &meta.StoreAddFriendResponse{Confirm: true}, nil
	}

	return &meta.StoreAddFriendResponse{}, nil
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
	// 再添加未推送消息队列
	if err := s.message.addQueue(req.Msg.ID); err != nil {
		return &meta.StoreNewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	// 再调用推送
	return nil, nil
}
