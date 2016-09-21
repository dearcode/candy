package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type Master struct {
	client meta.MasterClient
}

func NewMaster(host string) (*Master, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Master{client: meta.NewMasterClient(conn)}, nil
}

func (m *Master) NewID() (int64, error) {
	resp, err := m.client.NewID(context.Background(), &meta.NewIDRequest{})
	if err != nil {
		return 0, errors.Trace(err)
	}

	return resp.ID, errors.Trace(resp.Header.Error())
}
