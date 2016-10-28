package notice

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

type gateClient struct {
	clients map[string]meta.GateClient
	sync.Mutex
}

func newGateClient() *gateClient {
	return &gateClient{clients: make(map[string]meta.GateClient)}
}

func (g *gateClient) getClient(addr string) (meta.GateClient, error) {
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

func (g *gateClient) push(addr string, req meta.PushRequest) error {
	c, err := g.getClient(addr)
	if err != nil {
		return errors.Trace(err)
	}

	resp, err := c.Push(context.Background(), &req)
	if err != nil {
		return errors.Trace(err)
	}

	return resp.Header.Error()
}
