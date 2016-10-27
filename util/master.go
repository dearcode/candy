package util

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/ngaut/log"
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

	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &MasterClient{conn: conn, client: meta.NewMasterClient(conn), host: host, etcd: etcd}, nil
}

const (
	maxRetryTimes = 10
)

// BackOff Implements exponential backoff with full jitter.
// Returns real back off time in microsecond.
// See http://www.awsarchitectureblog.com/2015/03/backoff.html.
func BackOff(attempts int) int {
	upper := int(math.Min(float64(retryBackOffCap), float64(retryBackOffBase)*math.Pow(2.0, float64(attempts))))
	sleep := time.Duration(rand.Intn(upper)) * time.Millisecond
	time.Sleep(sleep)
	return int(sleep)
}

var (
	// Max retry count
	maxRetryCnt int = 10
	// retryBackOffBase is the initial duration
	retryBackOffBase = float64(1)
	// retryBackOffCap is the max amount of duration
	retryBackOffCap = float64(100)
)

type Retry struct {
	attempts   int
	times      time.Duration
	maxAttemps int
	maxTime    time.Duration
}

func NewRetry() *Retry {
	return &Retry{maxAttemps: 10}
}

func (r *Retry) Valid() bool {
	if r.maxAttemps != 0 && r.attempts >= r.maxAttemps {
		return false
	}

	if r.maxTime != 0 && r.times >= r.maxTime {
		return false
	}

	return true
}

func (r *Retry) Attempts() int {
	return r.attempts
}
func (r *Retry) Next() {
	delta := math.Min(retryBackOffCap, retryBackOffBase*math.Pow(2.0, float64(r.attempts)))
	sleep := time.Duration(rand.Float64()*delta) * time.Millisecond
	time.Sleep(sleep)
	r.times += sleep
	r.attempts++
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

	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
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
		break
		err = e
		if e := m.reconnectMaster(); e != nil {
			log.Errorf("reconnectMaster error:%s", e.Error())
			err = e
		}
	}

	m.Unlock()

	return id, err
}
