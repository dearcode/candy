package notice

import (
	"sync"

	"github.com/juju/errors"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
)

type gate struct {
	clients map[string]meta.GateClient
	sync.RWMutex
}

func newGate() *gate {
	return &gate{clients: make(map[string]meta.GateClient)}
}

func (g *gate) client(addr string) (meta.GateClient, error) {
	g.RLock()
	c, ok := g.clients[addr]
	g.RUnlock()
	if ok {
		return c, nil
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	c = meta.NewGateClient(conn)

	g.Lock()
	if _, ok := g.clients[addr]; ok {
		conn.Close()
	} else {
		g.clients[addr] = c
	}
	g.Unlock()

	return c, nil
}
