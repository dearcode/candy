package util

import (
	"encoding/json"
	"sync"

	"github.com/biogo/store/llrb"
	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

var (
	//ErrRegionNotFound region not exist.
	ErrRegionNotFound = errors.New("region not found")
)

//Regions 保存region的路由信息.
type Regions struct {
	mu    sync.RWMutex
	tree  *llrb.Tree
	hosts map[string]*meta.Region
}

//NewRegions 初始化llrb，map.
func NewRegions(ops ...RegionsOption) (*Regions, error) {
	r := &Regions{tree: &llrb.Tree{}, hosts: make(map[string]*meta.Region)}
	for _, op := range ops {
		if err := op(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// RegionsOption Regions option.
type RegionsOption func(*Regions) error

//RegionsWithHosts 初始化hosts
func RegionsWithHosts(hosts string) RegionsOption {
	return func(r *Regions) error {
		if err := json.Unmarshal([]byte(hosts), &r.hosts); err != nil {
			return err
		}
		for _, v := range r.hosts {
			r.tree.Insert(v)
		}
		return nil
	}
}

//Max 当前跨度最大的节点
func (r *Regions) Max() (meta.Region, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.tree.Len() == 0 {
		return meta.Region{}, ErrRegionNotFound
	}

	var o *meta.Region
	r.tree.Do(func(c llrb.Comparable) (done bool) {
		n := c.(*meta.Region)
		if o == nil || n.Span()-o.Span() >= o.Span()/2 {
			o = n
		}
		return
	})
	return *o, nil
}

//Split 添加一个节点并分割相对大的节点.
func (r *Regions) Split(from, to string) (meta.Region, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tree.Len() == 0 {
		n := meta.NewRegion(to, 0, meta.MaxRegionEnd)
		r.tree.Insert(n)
		r.hosts[to] = n
		return *n, nil
	}

	f, ok := r.hosts[from]
	if !ok {
		log.Infof("from %s not found", from)
		return meta.Region{}, ErrRegionNotFound
	}

	t, ok := r.hosts[to]
	if ok {
		log.Infof("new region %s exist", to)
		return *t, nil
	}

	delta := (f.End - f.Begin) / 2

	t = meta.NewRegion(to, f.End-delta, f.End)

	r.tree.Delete(f)
	f.End -= delta
	r.tree.Insert(f)

	r.tree.Insert(t)

	r.hosts[to] = t

	return *t, nil
}

//Get 根据ID获取对应的region信息
func (r *Regions) Get(id int64) (meta.Region, bool) {
	b := int32(id % meta.MaxRegionEnd)
	r.mu.RLock()
	c := r.tree.Get(&meta.Region{Begin: b})
	if c == nil {
		return meta.Region{}, false
	}
	src := c.(*meta.Region)
	r.mu.RUnlock()

	return *src, true
}

//GetByHost 根据host获取对应的region信息.
func (r *Regions) GetByHost(host string) (meta.Region, error) {
	r.mu.RLock()
	if src, ok := r.hosts[host]; ok {
		r.mu.RUnlock()
		return *src, nil
	}
	r.mu.RUnlock()

	return meta.Region{}, ErrRegionNotFound
}

// Next 返回当前节点的下一个节点
func (r *Regions) Next(host string) (self meta.Region, next meta.Region, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.hosts[host]
	if !ok {
		log.Infof("host:%s not found", host)
		err = ErrRegionNotFound
		return
	}

	if r.tree.Len() == 1 {
		log.Infof("only one region")
		err = ErrRegionNotFound
		return
	}

	begin := f.End
	if begin == meta.MaxRegionEnd {
		begin = f.Begin - 1
	}

	c := r.tree.Get(&meta.Region{Begin: begin})
	if c == nil {
		log.Infof("region begin:%d not found", begin)
		err = ErrRegionNotFound
		return
	}

	self = *f
	next = *c.(*meta.Region)

	return
}

// Merge 合并节点,合并当前删除的节点到前一个节点
func (r *Regions) Merge(from, to string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.hosts[from]
	if !ok {
		log.Infof("from %s not found", from)
		return ErrRegionNotFound
	}

	t, ok := r.hosts[to]
	if !ok {
		log.Infof("to %s not found", to)
		return ErrRegionNotFound
	}

	r.tree.Delete(f)
	r.tree.Delete(t)
	delete(r.hosts, from)

	t.Begin, t.End = f.Max(*t)

	r.tree.Insert(t)
	return nil
}

//Marshal 导出json格式数据
func (r *Regions) Marshal() (b []byte, e error) {
	r.mu.RLock()
	b, e = json.Marshal(r.hosts)
	r.mu.RUnlock()
	return
}

//Dump 导出region
func (r *Regions) Dump() []meta.Region {
	var l []meta.Region
	r.mu.RLock()
	for _, v := range r.hosts {
		l = append(l, *v)
	}
	r.mu.RUnlock()
	return l
}
