package util

import (
	"sync"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// MasterClient 连接master服务.
type MasterClient struct {
	host   string
	conn   *grpc.ClientConn
	client meta.MasterClient
	etcd   *EtcdClient
	sync.Mutex
}

// NewMasterClient 返回master client.
func NewMasterClient(host string, etcdAddrs []string) (*MasterClient, error) {
	var etcd *EtcdClient
	var err error

	if len(etcdAddrs) != 0 {
		if etcd, err = NewEtcdClient(etcdAddrs); err != nil {
			return nil, errors.Annotatef(err, "etcdAddrs size:%d", len(etcdAddrs))
		}

		if host, err = etcd.Get(EtcdMasterAddrKey); err != nil {
			return nil, errors.Trace(err)
		}
	}

	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout), grpc.WithBackoffMaxDelay(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &MasterClient{conn: conn, client: meta.NewMasterClient(conn), host: host, etcd: etcd}, nil
}

func (m *MasterClient) reconnect() error {
	log.Debugf("close old connect:%+v, from:%s", m.conn, m.host)
	m.conn.Close()

	var host string
	var err error

	if m.etcd != nil {
		if host, err = m.etcd.Get(EtcdMasterAddrKey); err != nil {
			return errors.Trace(err)
		}
	} else {
		host = m.host
	}

	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout), grpc.WithBackoffMaxDelay(NetworkTimeout))
	if err != nil {
		return errors.Trace(err)
	}

	m.host = host
	m.conn = conn
	m.client = meta.NewMasterClient(conn)

	return nil
}

//service 带重连的
func (m *MasterClient) service(call func(context.Context, meta.MasterClient) error) {
	m.Lock()
	for r := NewRetry(); r.Valid(); r.Next() {
		ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
		err := call(ctx, m.client)
		cancel()
		if err == nil {
			break
		}

		if err = m.reconnect(); err != nil {
			log.Errorf("reconnect error:%s", err.Error())
		}
	}
	m.Unlock()
}

// NewID 生成新ID.
func (m *MasterClient) NewID() (int64, error) {
	var err error
	req := &meta.NewIDRequest{}
	var resp *meta.NewIDResponse

	m.service(func(ctx context.Context, client meta.MasterClient) error {
		resp, err = client.NewID(ctx, req)
		return err
	})
	if err != nil {
		return 0, err
	}
	return resp.ID, err
}

// RegionGet 根据host获取region
func (m *MasterClient) RegionGet(host string) ([]meta.Region, error) {
	var resp *meta.RegionGetResponse
	var err error
	req := &meta.RegionGetRequest{Host: host}

	m.service(func(ctx context.Context, client meta.MasterClient) error {
		resp, err = client.RegionGet(ctx, req)
		return err
	})
	if err != nil {
		return nil, err
	}

	return resp.Regions, err
}
