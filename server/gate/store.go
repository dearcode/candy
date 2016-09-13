package gate

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
)

type store struct {
	host string
	ctx  context.Context
	api  meta.StoreClient
}

func newStore(host string) *store {
	return &store{host: host, ctx: context.Background()}
}

func (s *store) start() error {
	log.Debug("store start...")
	conn, err := grpc.Dial(s.host, grpc.WithInsecure(), grpc.WithTimeout(util.NetworkTimeout))
	if err != nil {
		return errors.Trace(err)
	}
	s.api = meta.NewStoreClient(conn)
	return nil
}

func (s *store) register(user, passwd string, id int64) error {
	log.Debugf("store register, user:%v passwd:%v", user, passwd)
	req := &meta.StoreRegisterRequest{User: user, Password: passwd, ID: id}
	resp, err := s.api.Register(s.ctx, req)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}

func (s *store) auth(user, passwd string) (int64, error) {
	log.Debugf("store auth, user:%v passwd:%v", user, passwd)
	req := &meta.StoreAuthRequest{User: user, Password: passwd}
	resp, err := s.api.Auth(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.ID, nil
}

func (s *store) updateUserInfo(user, nickName string, avatar []byte) (int64, error) {
	log.Debugf("updateUserInfo user:%v nickName:%v", user, nickName)
	req := &meta.StoreUpdateUserInfoRequest{User: user, NickName: nickName, Avatar: avatar}
	resp, err := s.api.UpdateUserInfo(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	log.Debugf("updateUserInfo success")
	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}
	log.Debugf("success")

	return resp.ID, nil
}

func (s *store) getUserInfo(user string) (int64, string, string, []byte, error) {
	log.Debugf("get userInfo user:%v", user)
	req := &meta.StoreGetUserInfoRequest{User: user}
	resp, err := s.api.GetUserInfo(s.ctx, req)
	if err != nil {
		return -1, "", "", nil, errors.Trace(err)
	}

	if resp.Header != nil {
		return -1, "", "", nil, errors.New(resp.Header.Msg)
	}
	log.Debugf("success")

	return resp.ID, resp.User, resp.NickName, resp.Avatar, nil
}

func (s *store) findUser(user string) (int64, error) {
	log.Debugf("store findUser, user:%v", user)
	req := &meta.StoreFindUserRequest{User: user}
	resp, err := s.api.FindUser(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.ID, nil
}

func (s *store) addFriend(from, to int64, confirm bool) (bool, error) {
	log.Debugf("store addFriend, from:%v to:%v confirm:%v", from, to, confirm)
	req := &meta.StoreAddFriendRequest{From: from, To: to, Confirm: confirm}
	resp, err := s.api.AddFriend(s.ctx, req)
	if err != nil {
		return false, errors.Trace(err)
	}

	if resp.Header != nil {
		return false, errors.New(resp.Header.Msg)
	}

	return resp.Confirm, nil
}

func (s *store) createGroup(userID, groupID int64) error {
	log.Debugf("store createGroup, userID:%v groupID:%v", userID, groupID)
	req := &meta.StoreCreateGroupRequest{UserID: userID, GroupID: groupID}
	resp, err := s.api.CreateGroup(s.ctx, req)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}
