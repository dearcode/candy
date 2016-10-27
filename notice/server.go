package notice

import (
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
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
}

// NewNotifer new Notifer server.
func NewNotifer(host string) *Notifer {
	return &Notifer{host: host, broker: newBroker()}
}

// Start start service.
func (n *Notifer) Start() error {
	log.Debug("notice Start...")
	serv := grpc.NewServer()

	meta.RegisterPushServer(serv, n)

	lis, err := net.Listen("tcp", n.host)
	if err != nil {
		return err
	}

	n.broker.start()

	return serv.Serve(lis)
}

// Subscribe add watch to chan or disable watch.
func (n *Notifer) Subscribe(stream meta.Push_SubscribeServer) error {
	addr, err := util.ContextAddr(stream.Context())
	if err != nil {
		return err
	}

	pushChan := make(chan pushRequest, defaultChanSize)
	cid := n.broker.addPushChan(pushChan)

	log.Debugf("gate:%s, chanID:%d", addr, cid)

	go func() {
		for req, ok := <-pushChan; ok; req, ok = <-pushChan {
			if err = stream.Send(&meta.PushRequest{ID: req.ids, Msg: req.msg}); err != nil {
				log.Errorf("stream send:(%v) error:%s", req, err.Error())
				//TODO 关闭stream
				return
			}
		}
		log.Errorf("pushChan close")
	}()

	for {
		req, err := stream.Recv()
		if err != nil {
			//TODO 关闭stream
			n.broker.delPushChan(cid)
			close(pushChan)
			log.Errorf("stream recv error:%s", err.Error())
			break
		}
		if req.Enable {
			n.broker.subscribe(req.ID, req.Device, cid)
			log.Debugf("subscribe user:%d dev:%s cid:%d", req.ID, req.Device, cid)
		} else {
			n.broker.unSubscribe(req.ID, req.Device)
			log.Debugf("unsubscribe user:%d dev:%s", req.ID, req.Device)
		}
	}

	return nil
}

// Push push a message to gate.
func (n *Notifer) Push(_ context.Context, req *meta.PushRequest) (*meta.PushResponse, error) {
	log.Debugf("begin push message:%v, ids:%v", req.Msg, req.ID)
	var ids []meta.PushID
	for _, id := range req.ID {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		log.Errorf("end push ids is nil, ids:%v, msg:%v", req.ID, req.Msg)
		return &meta.PushResponse{Header: &meta.ResponseHeader{Msg: "push ids is nil", Code: -1}}, nil
	}

	n.broker.push(req.Msg, ids...)
	log.Debugf("end push message:%v, ids:%v", req.Msg, req.ID)

	return &meta.PushResponse{}, nil
}
