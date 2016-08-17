package gate

import (
	"github.com/juju/errors"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type store struct {
	host string
	meta.StoreClient
}

func newStore(host string) *store {
	return &store{host: host}
}

func (s *store) start() error {
	conn, err := grpc.Dial(s.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return errors.Trace(err)
	}
	s.StoreClient = meta.NewStoreClient(conn)
	return nil
}

func (s *store) register(user, passwd string, id int64) error {
	req := &meta.RegisterRequest{User: user, Password: passwd, ID: id}
	resp, err := s.Register(nil, req)
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
	resp, err := s.Auth(nil, req)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if resp.Header != nil {
		return 0, errors.New(resp.Header.Msg)
	}

	return resp.ID, nil
}
