package notice

import (
	"net"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

const (
	networkTimeout  = time.Second * 3
	defaultChanSize = 1000
)

// Notifer recv client request.
type Notifer struct {
	host   string
	broker *broker
	gate   *gateClient
	serv   *grpc.Server
	ln     net.Listener
}

// NewNotifer new Notifer server.
func NewNotifer(host string) (*Notifer, error) {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	gate := newGateClient()

	n := &Notifer{
		host:   host,
		gate:   gate,
		broker: newBroker(gate),
		serv:   grpc.NewServer(),
		ln:     ln,
	}

	meta.RegisterNotiferServer(n.serv, n)

	return n, n.serv.Serve(ln)
}

// UnSubscribe 用户下线，取消在线推送
func (n *Notifer) UnSubscribe(_ context.Context, req *meta.UnSubscribeRequest) (*meta.UnSubscribeResponse, error) {
	n.broker.unSubscribe(req.ID, req.Token, req.Device)

	return &meta.UnSubscribeResponse{}, nil
}

// Subscribe 用户上线，接受在线推送
func (n *Notifer) Subscribe(_ context.Context, req *meta.SubscribeRequest) (*meta.SubscribeResponse, error) {
	n.broker.subscribe(req.ID, req.Token, req.Device, req.Host)
	return &meta.SubscribeResponse{}, nil
}

// Push store调用的接口.
func (n *Notifer) Push(_ context.Context, req *meta.PushRequest) (*meta.PushResponse, error) {
	log.Debugf("begin push message:%v, ids:%v", req.Msg, req.ID)
	if len(req.ID) == 0 {
		log.Errorf("end push ids is nil, ids:%v, msg:%v", req.ID, req.Msg)
		return &meta.PushResponse{Header: &meta.ResponseHeader{Msg: "push ids is nil", Code: -1}}, nil
	}

	n.broker.push(req.Msg, req.ID...)
	log.Debugf("end push message:%v, ids:%v", req.Msg, req.ID)

	return &meta.PushResponse{}, nil
}
