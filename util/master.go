package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

// Master 连接master服务.
type Master struct {
	client meta.MasterClient
}

// NewMaster 返回master client.
func NewMaster(host string) (*Master, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Master{client: meta.NewMasterClient(conn)}, nil
}

// NewID 生成新ID.
func (m *Master) NewID() (int64, error) {
	resp, err := m.client.NewID(context.Background(), &meta.NewIDRequest{})
	if err != nil {
		return 0, errors.Trace(err)
	}

	return resp.ID, errors.Trace(resp.Header.Error())
}
