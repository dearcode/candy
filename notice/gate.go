package notice

import (
	"net"
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type gateClient struct {
	clients map[string]meta.GateClient
	token   string
	ctx     context.Context
	sync.Mutex
}

func newGateClient() *gateClient {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}

	token := ""
	for _, i := range interfaces {
		if i.HardwareAddr != nil {
			token = i.HardwareAddr.String()
		}
	}

	log.Debugf("token:%s", token)
	ctx := util.ContextSet(context.Background(), "a", "b")
	ctx = util.ContextSet(ctx, "token", token)

	return &gateClient{clients: make(map[string]meta.GateClient), token: token, ctx: ctx}
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

	log.Debugf("ctx:%+v, req:%+v", g.ctx, req)
	resp, err := c.Push(g.ctx, &req)
	if err != nil {
		return errors.Trace(err)
	}

	return resp.Header.Error()
}
