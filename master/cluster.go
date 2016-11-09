package master

import (
	"sync"
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	nodeLost  = time.Second * 2
	regionKey = util.EtcdMasterKey + "/region"
)

type node struct {
	host string
	last time.Time
}

type cluster struct {
	host    string
	nodes   map[string]*node
	etcd    *util.EtcdClient
	closer  chan struct{}
	regions *util.Regions
	sync.RWMutex
}

func newCluster(etcd *util.EtcdClient, host string) (*cluster, error) {
	val, err := etcd.Get(regionKey)
	if err != nil {
		return nil, err
	}

	r, err := util.NewRegions(util.RegionsWithHosts(val))
	if err != nil {
		return nil, err
	}

	c := &cluster{nodes: make(map[string]*node), regions: r, etcd: etcd, host: host}

	go c.check()
	return c, nil
}

func (c *cluster) onHealth(host string) {
	c.Lock()
	if n, ok := c.nodes[host]; ok {
		n.last = time.Now()
	}
	c.Unlock()
}

func (c *cluster) get(host string) (meta.Region, error) {
	r, err := c.regions.GetByHost(host)
	if err == nil {
		return r, nil
	}

	if err != util.ErrRegionNotFound {
		return meta.Region{}, err
	}

	//添加节点
	return c.onNodeArrived(host)
}

func (c *cluster) getRegions() []meta.Region {
	return c.regions.Dump()
}

func (c *cluster) check() {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
		case <-c.closer:
			t.Stop()
			log.Debugf("stop")
			return
		}

		c.Lock()
		for _, n := range c.nodes {
			if time.Now().Sub(n.last) > nodeLost {
				c.onNodeExpired(n.host)
				delete(c.nodes, n.host)
			}
		}
		c.Unlock()
	}
}

func (c *cluster) stop() {
	if c.closer != nil {
		close(c.closer)
		c.closer = nil
	}
}

//onNodeExpire 节点超时
func (c *cluster) onNodeExpired(host string) error {
	self, next, err := c.regions.Next(host)
	if err != nil {
		log.Errorf("find next region error:%s", err.Error())
		return err
	}

	begin, end := self.Max(next)

	n, err := util.NewNotiferClient(next.Host)
	if err != nil {
		log.Errorf("connect notifer error:%s", err.Error())
		return err
	}
	defer n.Stop()

	//通知这个节点，修改自己负责的region范围
	if err = n.RegionSet(begin, end); err != nil {
		return err
	}

	if err = c.regions.Merge(host, next.Host); err != nil {
		return err
	}

	val, err := c.regions.Marshal()
	if err != nil {
		return err
	}

	if err = c.etcd.CAS(util.EtcdMasterAddrKey, c.host, regionKey, string(val)); err != nil {
		return err
	}

	c.Lock()
	delete(c.nodes, host)
	c.Unlock()

	return nil
}

//onNodeArrive 新节点上线, 失败不需要处理, 让客户端重试
func (c *cluster) onNodeArrived(host string) (meta.Region, error) {
	r, err := c.regions.Max()
	if err == util.ErrRegionNotFound {
		return c.regions.Split("", host)
	}

	notifer, err := util.NewNotiferClient(r.Host)
	if err != nil {
		log.Errorf("connect notifer error:%s", err.Error())
		return meta.Region{}, err
	}
	defer notifer.Stop()

	//通知原节点，修改范围
	if err = notifer.RegionSet(r.Begin, r.End/2); err != nil {
		return meta.Region{}, err
	}

	n, err := c.regions.Split(r.Host, host)
	if err != nil {
		return meta.Region{}, err
	}

	val, err := c.regions.Marshal()
	if err != nil {
		return meta.Region{}, err
	}

	if err = c.etcd.CAS(util.EtcdMasterAddrKey, c.host, regionKey, string(val)); err != nil {
		return meta.Region{}, err
	}

	c.Lock()
	c.nodes[host] = &node{host: host, last: time.Now()}
	c.Unlock()

	return n, nil
}

func (c *cluster) marshal() (string, error) {
	buf, err := c.regions.Marshal()
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
