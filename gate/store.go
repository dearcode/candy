package gate

import (
	"github.com/juju/errors"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type store struct {
	host string
	api  meta.StoreClient
}

func newStore(host string) *store {
	return &store{host: host}
}

func (s *store) start() error {
	conn, err := grpc.Dial(s.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return errors.Trace(err)
	}
	s.api = meta.NewStoreClient(conn)
	return nil
}

func (s *store) register(user, passwd string, id int64) error {
	req := &meta.RegisterRequest{User: user, Password: passwd, ID: id}
	resp, err := s.api.Register(nil, req)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}

func (s *store) auth(user, passwd string) (int64, error) {
	req := &meta.AuthRequest{User: user, Password: passwd}
	resp, err := s.api.Auth(nil, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.ID, nil
}

func (s *store) findUser(user string) (int64, error) {
	req := &meta.FindUserRequest{User: user}
	resp, err := s.api.FindUser(nil, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.ID, nil
}

func (s *store) addFriend(from, to int64, confirm bool) (bool, error) {
	req := &meta.AddFriendRequest{From: from, To: to, Confrim: confirm}
	resp, err := s.api.AddFriend(nil, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.Confrim, nil
}

func (s *store) createGroup(userID, groupID int64) error {
	req := &meta.CreateGroupRequest{UserID: userID, GroupID: groupID}
	resp, err := s.api.CreateGroup(nil, req)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}
