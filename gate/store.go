package gate

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
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

	log.Debugf("register finished, user:%s, id:%d, err:%v", user, id, resp.Header.Error())
	return errors.Trace(resp.Header.Error())
}

func (s *store) auth(user, passwd string) (int64, error) {
	log.Debugf("store auth, user:%v passwd:%v", user, passwd)
	req := &meta.StoreAuthRequest{User: user, Password: passwd}
	resp, err := s.api.Auth(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	log.Debugf("auth finished, user:%s, id:%d, err:%v", user, resp.ID, resp.Header.Error())
	return resp.ID, errors.Trace(resp.Header.Error())
}

func (s *store) updateUserInfo(user, nickName string, avatar []byte) (int64, error) {
	log.Debugf("updateUserInfo user:%v nickName:%v", user, nickName)
	req := &meta.StoreUpdateUserInfoRequest{User: user, NickName: nickName, Avatar: avatar}
	resp, err := s.api.UpdateUserInfo(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	log.Debugf("updateUserInfo finished, id:%d, err:%v", resp.ID, resp.Header.Error())
	return resp.ID, errors.Trace(resp.Header.Error())
}

func (s *store) updateUserPassword(user, passwd string) (int64, error) {
	log.Debugf("updateUserPassword user:%v passwd:%v", user, passwd)
	req := &meta.StoreUpdateUserPasswordRequest{User: user, Password: passwd}
	resp, err := s.api.UpdateUserPassword(s.ctx, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	log.Debugf("updateUserPassword finished, id:%d, err:%v", resp.ID, resp.Header.Error())
	return resp.ID, errors.Trace(resp.Header.Error())
}

func (s *store) getUserInfo(user string) (int64, string, string, []byte, error) {
	log.Debugf("get userInfo user:%v", user)
	req := &meta.StoreGetUserInfoRequest{User: user}
	resp, err := s.api.GetUserInfo(s.ctx, req)
	if err != nil {
		return -1, "", "", nil, errors.Trace(err)
	}

	log.Debugf("get userInfo finished, user:%s, err:%v", user, resp.Header.Error())

	return resp.ID, resp.User, resp.NickName, resp.Avatar, errors.Trace(resp.Header.Error())
}

func (s *store) findUser(user string) ([]string, error) {
	log.Debugf("store findUser, user:%v", user)
	req := &meta.StoreFindUserRequest{User: user}
	resp, err := s.api.FindUser(s.ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resp.Users, errors.Trace(resp.Header.Error())
}

func (s *store) addFriend(from, to int64, state meta.FriendRelation, msg string) (meta.FriendRelation, error) {
	log.Debugf("store addFriend, from:%v to:%v state:%v", from, to, state)
	req := &meta.StoreAddFriendRequest{From: from, To: to, State: state, Msg: msg}
	resp, err := s.api.AddFriend(s.ctx, req)
	if err != nil {
		return meta.FriendRelation_None, errors.Trace(err)
	}

	return resp.State, errors.Trace(resp.Header.Error())
}

func (s *store) createGroup(userID, groupID int64, name string) error {
	log.Debugf("store createGroup, userID:%v groupID:%v groupName:%v", userID, groupID, name)
	req := &meta.StoreCreateGroupRequest{UserID: userID, GroupID: groupID, GroupName: name}
	resp, err := s.api.CreateGroup(s.ctx, req)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(resp.Header.Error())
}

func (s *store) uploadFile(userID int64, data []byte) error {
	log.Debugf("store UploadFile, userID:%v", userID)
	req := &meta.StoreUploadFileRequest{
		Header: &meta.RequestHeader{User: userID},
		File:   data,
	}
	resp, err := s.api.UploadFile(s.ctx, req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("store UploadFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return errors.Trace(resp.Header.Error())
}

func (s *store) checkFile(userID int64, names []string) ([]string, error) {
	log.Debugf("store checkFile, userID:%v", userID)
	req := &meta.StoreCheckFileRequest{
		Header: &meta.RequestHeader{User: userID},
		Names:  names,
	}

	resp, err := s.api.CheckFile(s.ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Debugf("store checkFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return resp.Names, errors.Trace(resp.Header.Error())
}

func (s *store) downloadFile(userID int64, names []string) (map[string][]byte, error) {
	log.Debugf("store downloadFile, userID:%v", userID)
	req := &meta.StoreDownloadFileRequest{
		Header: &meta.RequestHeader{User: userID},
		Names:  names,
	}

	resp, err := s.api.DownloadFile(s.ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Debugf("store downloadFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return resp.Files, errors.Trace(resp.Header.Error())
}

func (s *store) newMessage(msg *meta.Message) error {
	log.Debugf("store newMessage")

	req := &meta.StoreNewMessageRequest{Msg: msg}
	resp, err := s.api.NewMessage(s.ctx, req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("send message success, resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}

func (s *store) loadMessage(userID, id int64, reverse bool) ([]*meta.Message, error) {
	log.Debugf("store loadMessage")
	req := &meta.StoreLoadMessageRequest{User: userID, ID: id, Reverse: reverse}
	resp, err := s.api.LoadMessage(s.ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	log.Debugf("send message success, resp:%v", resp)
	return resp.Msgs, errors.Trace(resp.Header.Error())
}
