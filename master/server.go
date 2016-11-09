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
	cluster   *cluster

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
		go m.runWithLeader()
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
		c, err := newCluster(m.etcd, m.host)
		if err != nil {
			log.Errorf("newCluster error:%s", errors.ErrorStack(err))
			return errors.Trace(err)
		}

		m.cluster = c
		m.allocator = a
	} else {
		if m.allocator != nil {
			m.allocator.stop()
			m.allocator = nil
		}
		if m.cluster != nil {
			m.cluster.stop()
			m.cluster = nil
		}
	}

	m.Unlock()
	return nil
}

func (m *Master) runWithLeader() {
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

// RegionGet 获取自己节点的region
func (m *Master) RegionGet(_ context.Context, req *meta.RegionGetRequest) (*meta.RegionGetResponse, error) {
	if m.cluster == nil {
		return &meta.RegionGetResponse{Regions: []meta.Region{{0, meta.MaxRegionEnd, ""}}}, nil
	}

	if req.Host != "" {
		r, err := m.cluster.get(req.Host)
		if err != nil {
			return &meta.RegionGetResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
		}
		return &meta.RegionGetResponse{Regions: []meta.Region{r}}, nil
	}

	return &meta.RegionGetResponse{Regions: m.cluster.getRegions()}, nil
}
