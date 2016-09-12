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
func (s *Store) Register(_ context.Context, req *meta.RegisterRequest) (*meta.RegisterResponse, error) {
	log.Debugf("Store Register, user:%v passwd:%v ID:%v", req.User, req.Password, req.ID)
	if err := s.user.register(req.User, req.Password, req.ID); err != nil {
		return &meta.RegisterResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.RegisterResponse{}, nil
}

// UpdateInfo update user base info, ex: nickname, picurl and so on
func (s *Store) UpdateInfo(_ context.Context, req *meta.UpdateInfoRequest) (*meta.UpdateInfoResponse, error) {
	log.Debugf("Store UpdateInfo, user:%v niceName:%v", req.User, req.NickName)
	if err := s.user.updateUserInfo(req.User, req.NickName, req.Avatar); err != nil {
		return &meta.UpdateInfoResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.UpdateInfoResponse{}, nil
}

// UpdatePassword update user password
func (s *Store) UpdatePassword(_ context.Context, req *meta.UpdatePasswordRequest) (*meta.UpdatePasswordResponse, error) {
	log.Debugf("Store UpdatePassword, user:")
	if err := s.user.updateUserPassword(req.User, req.Password); err != nil {
		return &meta.UpdatePasswordResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.UpdatePasswordResponse{}, nil
}

// Auth check password.
func (s *Store) Auth(_ context.Context, req *meta.AuthRequest) (*meta.AuthResponse, error) {
	log.Debugf("Store Auth, user:%v passwd:%v", req.User, req.Password)
	id, err := s.user.auth(req.User, req.Password)
	if err != nil {
		return &meta.AuthResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.AuthResponse{ID: id}, nil
}

// FindUser 根据字符串的用户名查的用户信息.
func (s *Store) FindUser(_ context.Context, req *meta.FindUserRequest) (*meta.FindUserResponse, error) {
	log.Debugf("Store FindUser, user:%v", req.User)
	id, err := s.user.findUser(req.User)
	if err != nil {
		return &meta.FindUserResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.FindUserResponse{ID: id}, nil
}

// AddFriend 添加好友，两人都添加过对方后才可以聊天.
func (s *Store) AddFriend(_ context.Context, req *meta.AddFriendRequest) (*meta.AddFriendResponse, error) {
	log.Debugf("Store AddFriend, from:%v to:%v Confirm:%v", req.From, req.To, req.Confirm)
	if err := s.user.friend.add(req.From, req.To, req.Confirm); err != nil {
		return &meta.AddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	if req.Confirm {
		if err := s.user.friend.add(req.To, req.From, req.Confirm); err != nil {
			return &meta.AddFriendResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}
		return &meta.AddFriendResponse{Confirm: true}, nil
	}

	return &meta.AddFriendResponse{}, nil
}

// CreateGroup create group.
func (s *Store) CreateGroup(_ context.Context, req *meta.CreateGroupRequest) (*meta.CreateGroupResponse, error) {
	log.Debugf("Store CreateGroup")
	return nil, nil
}

// NewMessage save message to leveldb,
func (s *Store) NewMessage(_ context.Context, req *meta.NewMessageRequest) (*meta.NewMessageResponse, error) {
	log.Debugf("Store NewMessage, msg:%v", req.Msg)
	// add消息到db
	if err := s.message.add(*req.Msg); err != nil {
		return &meta.NewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	// 再添加未推送消息队列
	if err := s.message.addQueue(req.Msg.ID); err != nil {
		return &meta.NewMessageResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	// 再调用推送
	return nil, nil
}
