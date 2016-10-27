package master

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/juju/errors"

	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	//EtcdIDKey Etcd ID Key
	EtcdIDKey = util.EtcdMasterKey + "/id"
)

type store interface {
	Get(string) (string, error)
	CAS(string, string, string, string) error
}

type allocator struct {
	store  store
	host   string
	isStop bool
	last   int64
	seq    int64
	sync.RWMutex
}

func newAllocator(store store, host string) (*allocator, error) {
	v, err := store.Get(EtcdIDKey)
	if err != nil {
		return nil, errors.Trace(err)
	}

	last, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for now := time.Now().Unix(); now <= last; now = time.Now().Unix() {
		log.Infof("now:%d must be bigger then last:%d", now, last)
		time.Sleep(time.Second)
	}

	a := &allocator{store: store, last: last, host: host}

	go a.save()

	log.Debugf("allocator start")

	return a, nil
}

func (a *allocator) id() int64 {
	var i int64

	log.Debugf("a:%+v", a)
	a.Lock()

	for i = time.Now().Unix(); i < a.last; i = time.Now().Unix() {
		log.Infof("now:%d < last:%d", i, a.last)
		time.Sleep(time.Second)
	}

	if i != a.last || a.seq >= math.MaxInt32 {
		a.last = i
		a.seq = 0
	}

	a.seq++

	i = a.last<<32 + a.seq

	a.Unlock()

	return i
}

func (a *allocator) save() {
	var id int64

	t := time.NewTicker(time.Second)
	defer t.Stop()

	a.RLock()
	last, stop := a.last, a.isStop
	a.RUnlock()

	for !stop {
		<-t.C

		if last != id {
			err := a.store.CAS(util.EtcdMasterAddrKey, a.host, EtcdIDKey, strconv.FormatInt(last, 10))
			if err != nil {
				log.Errorf("save last:%d error:%s", last, errors.ErrorStack(err))
			}
			id = last
		}

		a.RLock()
		last, stop = a.last, a.isStop
		a.RUnlock()
	}

}

func (a *allocator) stop() {
	a.Lock()
	a.isStop = true
	a.Unlock()
}
