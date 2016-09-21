package notice

import (
	"sync"

	"github.com/juju/errors"
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

func (g *gate) client(addr string) (meta.GateClient, error) {
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
