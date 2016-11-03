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

func (m *MasterClient) reconnectMaster() error {
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

// NewID 生成新ID.
func (m *MasterClient) NewID() (int64, error) {
	var id int64
	var err error

	m.Lock()
	for r := NewRetry(); r.Valid(); r.Next() {
		resp, e := m.client.NewID(context.Background(), &meta.NewIDRequest{})
		if e == nil {
			id = resp.ID
			err = resp.Header.Error()
			break
		}
		log.Errorf("call NewID error:%s, attempts:%d", e.Error(), r.Attempts())
		err = e
		if e := m.reconnectMaster(); e != nil {
			log.Errorf("reconnectMaster error:%s", e.Error())
			err = e
		}
	}

	m.Unlock()

	return id, err
}
