package notice

import (
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

const (
	networkTimeout = time.Second * 3
)

// Notifer recv client request.
type Notifer struct {
	host   string
	broker *broker
}

// NewNotifer new Notifer server.
func NewNotifer(host string) *Notifer {
	return &Notifer{host: host, broker: newBroker(newGate())}
}

// Start start service.
func (n *Notifer) Start() error {
	log.Debug("notice Start...")
	serv := grpc.NewServer()

	meta.RegisterNoticeServiceServer(serv, n)

	lis, err := net.Listen("tcp", n.host)
	if err != nil {
		return err
	}

	n.broker.Start()

	return serv.Serve(lis)
}

// Subscribe subscribe a Notifer.
func (n *Notifer) Subscribe(c context.Context, req *meta.SubscribeRequest) (*meta.SubscribeResponse, error) {
	n.broker.Subscribe(req.ID, req.Host)
	return &meta.SubscribeResponse{}, nil
}

// UnSubscribe unsubscribe a Notifer.
func (n *Notifer) UnSubscribe(_ context.Context, req *meta.UnSubscribeRequest) (*meta.UnSubscribeResponse, error) {
	n.broker.UnSubscribe(req.ID)
	return &meta.UnSubscribeResponse{}, nil
}

// Push push a message to gate.
func (n *Notifer) Push(_ context.Context, req *meta.PushRequest) (*meta.PushResponse, error) {
	log.Debugf("begin push message:%v, ids:%v", req.Msg, req.ID)
	n.broker.Push(*req.Msg, req.ID...)
	log.Debugf("end push message:%v, ids:%v", req.Msg, req.ID)

	return &meta.PushResponse{}, nil
}
