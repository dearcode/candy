package gate

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type storeClient struct {
	client meta.StoreClient
}

func newStoreClient(host string) (*storeClient, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(util.NetworkTimeout), grpc.WithBackoffMaxDelay(util.NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &storeClient{client: meta.NewStoreClient(conn)}, nil
}

func (s *storeClient) register(user, passwd string, id int64) error {
	log.Debugf("store register, user:%v passwd:%v", user, passwd)
	req := &meta.StoreRegisterRequest{
		User:     user,
		Password: passwd,
		ID:       id,
	}
	resp, err := s.client.Register(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("register finished, user:%s, id:%d, err:%v", user, id, resp.Header.Error())
	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) auth(user, passwd string) (int64, error) {
	log.Debugf("store auth, user:%v passwd:%v", user, passwd)
	req := &meta.StoreAuthRequest{
		User:     user,
		Password: passwd,
	}
	resp, err := s.client.Auth(context.Background(), req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	log.Debugf("auth finished, user:%s, id:%d, err:%v", user, resp.ID, resp.Header.Error())
	return resp.ID, errors.Trace(resp.Header.Error())
}

func (s *storeClient) updateUserInfo(id int64, name, nickName, avatar string) error {
	log.Debugf("%d name:%v nickName:%v", id, name, nickName)
	req := &meta.StoreUpdateUserInfoRequest{
		RequestHeader: meta.RequestHeader{User: id},
		Name:          name,
		NickName:      nickName,
		Avatar:        avatar,
	}
	resp, err := s.client.UpdateUserInfo(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("%d finished, id:%d, err:%v", id, resp.Header.Error())
	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) updateSignature(id int64, name, signature string) error {
	req := &meta.StoreUpdateSignatureRequest{
		RequestHeader: meta.RequestHeader{User: id},
		Name:          name,
		Signature:     signature,
	}
	resp, err := s.client.UpdateSignature(context.Background(), req)
	if err != nil {
		return errors.Trace(resp.Header.Error())
	}

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) updateUserPassword(id int64, name, passwd, newPasswd string) error {
	log.Debugf("%d name:%v passwd old:%v, new:%v", id, name, passwd, newPasswd)
	req := &meta.StoreUpdateUserPasswordRequest{
		RequestHeader: meta.RequestHeader{User: id},
		Name:          name,
		Password:      passwd,
		NewPassword:   newPasswd,
	}
	resp, err := s.client.UpdateUserPassword(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("%d name:%v error:%s", id, name, resp.Header.Error())
	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) getUserInfo(id int64, findByName bool, userName string, userID int64) (*meta.UserInfo, error) {
	log.Debugf("get userInfo findByName:%v userName:%v userID:%v", findByName, userName, userID)
	req := &meta.StoreGetUserInfoRequest{
		RequestHeader: meta.RequestHeader{User: id},
		ByName:        findByName,
		Name:          userName,
		ID:            userID,
	}
	resp, err := s.client.GetUserInfo(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	log.Debugf("get userInfo finished, user:%s, err:%v", userName, resp.Header.Error())
	userInfo := &meta.UserInfo{
		ID:        resp.ID,
		Name:      resp.User,
		NickName:  resp.NickName,
		Avatar:    resp.Avatar,
		Signature: resp.Signature,
	}

	return userInfo, errors.Trace(resp.Header.Error())
}

func (s *storeClient) findUser(id int64, user string) ([]string, error) {
	log.Debugf("store findUser, user:%v", user)
	req := &meta.StoreFindUserRequest{RequestHeader: meta.RequestHeader{User: id}, Name: user}
	resp, err := s.client.FindUser(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resp.Users, errors.Trace(resp.Header.Error())
}

func (s *storeClient) friend(id, to int64, operate meta.Relation, msg string) error {
	log.Debugf("store friend, from:%v to:%v operate:%v", id, to, operate)
	req := &meta.StoreFriendRequest{
		RequestHeader: meta.RequestHeader{User: id},
		From:          id,
		To:            to,
		Operate:       operate,
		Msg:           msg,
	}
	resp, err := s.client.Friend(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) loadFriendList(id int64) ([]int64, error) {
	req := &meta.StoreLoadFriendListRequest{RequestHeader: meta.RequestHeader{User: id}}
	resp, err := s.client.LoadFriendList(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resp.Users, errors.Trace(resp.Header.Error())
}

func (s *storeClient) groupCreate(uid, gid int64, name string) error {
	log.Debugf("store group create user:%v group:%v name:%v", uid, gid, name)
	req := &meta.StoreGroupCreateRequest{
		RequestHeader: meta.RequestHeader{User: uid},
		ID:            gid,
		Name:          name,
	}
	resp, err := s.client.GroupCreate(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) group(uid, gid int64, operate meta.Relation, users []int64, msg string) error {
	log.Debugf("store group delete, user:%v group:%v", uid, gid)
	req := &meta.StoreGroupRequest{
		RequestHeader: meta.RequestHeader{User: uid},
		ID:            gid,
		Operate:       operate,
		Users:         users,
		Msg:           msg,
	}
	resp, err := s.client.Group(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) groupDelete(uid, gid int64) error {
	log.Debugf("store group delete, user:%v group:%v", uid, gid)
	req := &meta.StoreGroupDeleteRequest{
		RequestHeader: meta.RequestHeader{User: uid},
		ID:            gid,
	}
	resp, err := s.client.GroupDelete(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) loadGroupList(id int64) ([]*meta.GroupInfo, error) {
	req := &meta.StoreLoadGroupListRequest{RequestHeader: meta.RequestHeader{User: id}}
	resp, err := s.client.LoadGroupList(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return resp.Groups, errors.Trace(resp.Header.Error())
}

func (s *storeClient) uploadFile(userID int64, data []byte) error {
	log.Debugf("store UploadFile, userID:%v", userID)
	req := &meta.StoreUploadFileRequest{
		RequestHeader: meta.RequestHeader{User: userID},
		File:          data,
	}
	resp, err := s.client.UploadFile(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("store UploadFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) checkFile(userID int64, names []string) ([]string, error) {
	log.Debugf("store checkFile, userID:%v", userID)
	req := &meta.StoreCheckFileRequest{
		RequestHeader: meta.RequestHeader{User: userID},
		Names:         names,
	}

	resp, err := s.client.CheckFile(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Debugf("store checkFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return resp.Names, errors.Trace(resp.Header.Error())
}

func (s *storeClient) downloadFile(userID int64, names []string) (map[string][]byte, error) {
	log.Debugf("store downloadFile, userID:%v", userID)
	req := &meta.StoreDownloadFileRequest{
		RequestHeader: meta.RequestHeader{User: userID},
		Names:         names,
	}

	resp, err := s.client.DownloadFile(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Debugf("store downloadFile finished, userID:%v, err:%v", userID, resp.Header.Error())

	return resp.Files, errors.Trace(resp.Header.Error())
}

func (s *storeClient) newMessage(id int64, msg meta.Message) error {
	log.Debugf("store newMessage[%d]:%+v", msg.ID, msg)

	req := &meta.StoreNewMessageRequest{
		RequestHeader: meta.RequestHeader{User: id},
		Msg:           msg,
	}
	resp, err := s.client.NewMessage(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("send message[%d] success, resp:%v", msg.ID, resp)
	return errors.Trace(resp.Header.Error())
}

func (s *storeClient) loadMessage(userID, id int64, reverse bool) ([]*meta.PushMessage, error) {
	log.Debugf("store loadMessage")
	req := &meta.StoreLoadMessageRequest{
		RequestHeader: meta.RequestHeader{User: userID},
		ID:            id,
		Reverse:       reverse,
	}
	resp, err := s.client.LoadMessage(context.Background(), req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	log.Debugf("send message success, resp:%v", resp)
	return resp.Msgs, errors.Trace(resp.Header.Error())
}
