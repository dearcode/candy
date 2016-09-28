package notice

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type gate struct {
	clients map[string]meta.GateClient
	sync.Mutex
}

func newGate() *gate {
	return &gate{clients: make(map[string]meta.GateClient)}
}

func (g *gate) getClient(addr string) (meta.GateClient, error) {
	g.Lock()
	defer g.Unlock()

	c, ok := g.clients[addr]
	if ok {
		return c, nil
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}

	c = meta.NewGateClient(conn)
	g.clients[addr] = c
	return c, nil
}

func (g *gate) notice(addr string, ids []*meta.PushID, msg *meta.Message) error {
	c, err := g.getClient(addr)
	if err != nil {
		return errors.Trace(err)
	}
	req := &meta.GateNoticeRequest{ID: ids, Msg: msg}
	resp, err := c.Notice(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	return resp.Header.Error()
}
