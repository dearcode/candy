package notice

import (
	"net"
	"sync"
	"time"

	"github.com/juju/errors"
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
	master *util.MasterClient
	gate   *gateClient
	serv   *grpc.Server
	region meta.Region
	ln     net.Listener

	sync.RWMutex
}

// NewServer new Notifer server.
func NewServer(host, master, etcd string) (*Notifer, error) {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	m, err := util.NewMasterClient(master, util.Split(etcd, ","))
	if err != nil {
		return nil, errors.Trace(err)
	}

	rs, err := m.RegionGet(host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	gate := newGateClient()

	n := &Notifer{
		host:   host,
		gate:   gate,
		master: m,
		region: rs[0],
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
	if !n.region.Match(req.ID) {
		log.Errorf("does not match id:%d, region:%+v", req.ID, n.region)
		return &meta.SubscribeResponse{Header: &meta.ResponseHeader{Msg: "does not match", Code: -2}}, nil
	}
	return &meta.SubscribeResponse{}, nil
}

// Push store调用的接口.
func (n *Notifer) Push(_ context.Context, req *meta.PushRequest) (*meta.PushResponse, error) {
	log.Debugf("begin push message:%v, ids:%v", req.Msg, req.ID)
	if len(req.ID) == 0 {
		log.Errorf("end push ids is nil, ids:%v, msg:%v", req.ID, req.Msg)
		return &meta.PushResponse{Header: &meta.ResponseHeader{Msg: "push ids is nil", Code: -1}}, nil
	}

	for _, id := range req.ID {
		if !n.region.Match(id.User) {
			log.Errorf("does not match id:%d, region:%+v", id.User, n.region)
			return &meta.PushResponse{Header: &meta.ResponseHeader{Msg: "does not match", Code: -2}}, nil
		}
	}

	n.broker.push(req.Msg, req.ID...)
	log.Debugf("end push message:%v, ids:%v", req.Msg, req.ID)

	return &meta.PushResponse{}, nil
}

// RegionSet 修改当前region的范围
func (n *Notifer) RegionSet(_ context.Context, req *meta.RegionSetRequest) (*meta.RegionSetResponse, error) {
	n.Lock()
	n.region.Begin = req.Begin
	n.region.End = req.End
	n.Unlock()

	return &meta.RegionSetResponse{}, nil
}
