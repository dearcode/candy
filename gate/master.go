package gate

import (
	"github.com/juju/errors"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type master struct {
	host string
	meta.MasterClient
}

func newMaster(host string) *master {
	return &master{host: host}
}

func (m *master) start() error {
	conn, err := grpc.Dial(m.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return errors.Trace(err)
	}
	m.MasterClient = meta.NewMasterClient(conn)
	return nil
}

func (m *master) newID() (int64, error) {
	resp, err := m.NewID(nil, &meta.NewIDRequest{})
	if err != nil {
		return 0, errors.Trace(err)
	}
	return resp.ID, nil
}
