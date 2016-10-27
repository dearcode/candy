package master

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

// Master process gate request.
type Master struct {
	serv      *grpc.Server
	etcd      *util.EtcdClient
	isLeader  bool
	host      string
	allocator *allocator

	closer chan struct{}

	sync.RWMutex
}

// NewMaster create new Master.
func NewMaster(host string, etcdAddrs []string) (*Master, error) {
	var etcd *util.EtcdClient
	var err error

	if len(etcdAddrs) != 0 {
		if etcd, err = util.NewEtcdClient(etcdAddrs); err != nil {
			return nil, errors.Trace(err)
		}
	}

	l, err := net.Listen("tcp", host)
	if err != nil {
		return nil, errors.Trace(err)
	}

	m := &Master{host: host, etcd: etcd, closer: make(chan struct{}), serv: grpc.NewServer()}

	meta.RegisterMasterServer(m.serv, m)

	if etcd != nil {
		go m.receiver()
	} else {
		m.allocator, _ = newAllocator(newMstore(), m.host)
	}

	return m, m.serv.Serve(l)
}

func (m *Master) service() error {
	m.Lock()
	if m.isLeader {
		a, err := newAllocator(m.etcd, m.host)
		if err != nil {
			log.Errorf("newAllocator error:%s", errors.ErrorStack(err))
			return errors.Trace(err)
		}
		m.allocator = a
	} else {
		if m.allocator != nil {
			m.allocator.stop()
			m.allocator = nil
		}
	}

	m.Unlock()
	return nil
}

func (m *Master) receiver() {
	state, stop := m.etcd.CampaignLeader(util.EtcdMasterAddrKey, m.host)

	for {
		select {
		case isLeader := <-state:
			m.Lock()
			m.isLeader = isLeader
			m.Unlock()
			log.Debugf("i am leader?:%v", isLeader)

			if err := m.service(); err != nil {
				panic(err.Error())
			}

		case <-m.closer:
			log.Infof("stop service, stop campaign leader")
			close(stop)
			return
		}
	}

}

// NewID return an new id
func (m *Master) NewID(_ context.Context, _ *meta.NewIDRequest) (*meta.NewIDResponse, error) {
	m.RLock()
	a := m.allocator
	m.RUnlock()

	if a == nil {
		return &meta.NewIDResponse{Header: &meta.ResponseHeader{Code: -1, Msg: "not leader"}}, nil
	}

	return &meta.NewIDResponse{ID: a.id()}, nil
}
