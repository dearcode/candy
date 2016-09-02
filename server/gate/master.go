package gate

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
)

type master struct {
	host string
	ctx  context.Context
	meta.MasterClient
}

func newMaster(host string) *master {
	return &master{host: host, ctx: context.Background()}
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
	resp, err := m.NewID(m.ctx, &meta.NewIDRequest{})
	if err != nil {
		return 0, errors.Trace(err)
	}
	return resp.ID, nil
}
