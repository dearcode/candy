package util

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"golang.org/x/net/context"
)

const (
	// EtcdMasterKey master key prefix.
	EtcdMasterKey = "/master"
	// EtcdMasterAddrKey master addr key.
	EtcdMasterAddrKey = EtcdMasterKey + "/addr"

	// 1秒超时
	etcdKeyTTL = 1
)

// EtcdClient etcd client.
type EtcdClient struct {
	client *clientv3.Client
}

var (
	//ErrMasterNotFound 未找到master.
	ErrMasterNotFound = errors.New("etcd master host not found")
)

// NewEtcdClient new etcd client.
func NewEtcdClient(addrs []string) (*EtcdClient, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: NetworkTimeout,
	})
	if err != nil {
		return nil, errors.Annotatef(err, "addrs:%+v", addrs)
	}

	return &EtcdClient{client: c}, nil
}

// Get get value from etcd.
func (e *EtcdClient) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	resp, err := clientv3.NewKV(e.client).Get(ctx, key)
	cancel()
	if err != nil {
		return "", errors.Trace(err)
	}

	if len(resp.Kvs) == 0 {
		log.Debugf("key:%s value not found", key)
		return "0", nil
	}

	log.Debugf("find key:%s, value:%s", key, string(resp.Kvs[0].Value))
	return string(resp.Kvs[0].Value), nil
}

// CAS put value to etcd.
func (e *EtcdClient) CAS(cmpKey, cmpValue, key, value string) error {
	cmp := clientv3.Compare(clientv3.Value(cmpKey), "=", cmpValue)
	if cmpValue == "" {
		cmp = clientv3.Compare(clientv3.CreateRevision(cmpKey), "=", 0)
	}
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	pr, err := e.client.Txn(ctx).
		If(cmp).
		Then(clientv3.OpPut(key, value)).
		Commit()
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	if !pr.Succeeded {
		return errors.New("put key failed")
	}

	return nil
}

func (e *EtcdClient) watch(key string, event mvccpb.Event_EventType) {
	watcher := clientv3.NewWatcher(e.client)
	defer watcher.Close()

	for {
		watchChan := watcher.Watch(e.client.Ctx(), key)
		for resp := range watchChan {
			if resp.Canceled {
				return
			}
			for _, ev := range resp.Events {
				if ev.Type == event {
					return
				}
			}
		}
	}
}

//WaitKeyDelete 等待Key被删除
func (e *EtcdClient) WaitKeyDelete(key string) {
	e.watch(key, mvccpb.DELETE)
}

//WaitKeyPut 等待Key被修改
func (e *EtcdClient) WaitKeyPut(key string) {
	e.watch(key, mvccpb.PUT)
}

// campaign 竞争leader，只有contxt关闭才会返回nil
func (e *EtcdClient) campaign(key, value string, state chan<- bool, stop <-chan struct{}) error {
	lessor := clientv3.NewLease(e.client)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	lr, err := lessor.Grant(ctx, etcdKeyTTL)
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), NetworkTimeout)
	pr, err := e.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, value, clientv3.WithLease(clientv3.LeaseID(lr.ID)))).
		Commit()
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	if !pr.Succeeded {
		return errors.New("campaign leader failed, other server may campaign ok")
	}

	state <- true
	log.Debugf("campaign success")

	ctx, cancel = context.WithCancel(context.Background())
	if _, err = lessor.KeepAlive(ctx, clientv3.LeaseID(lr.ID)); err != nil {
		return errors.Trace(err)
	}
	<-stop
	log.Debugf("campaign stop")
	cancel()
	return nil
}

// Close 关闭客户端
func (e *EtcdClient) Close() {
	e.client.Close()
}

// CampaignLeader 竞争leader
// 返回状态只读的chan与关闭只写的chan
// state如果状态变化则通知, stop外部关闭则停止keepalive
func (e *EtcdClient) CampaignLeader(key, value string) (<-chan bool, chan<- struct{}) {
	stop := make(chan struct{})
	state := make(chan bool, 1)
	go func(state chan<- bool, stop <-chan struct{}) {
		state <- false
		for err := e.campaign(key, value, state, stop); err != nil; err = e.campaign(key, value, state, stop) {
			log.Errorf("campaignLeader error:%s, key:%s", errors.ErrorStack(err), key)
			e.WaitKeyDelete(key)
		}
		close(state)
		log.Debugf("CampaignLeader over")
	}(state, stop)
	return state, stop
}
