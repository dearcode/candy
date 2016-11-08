package util

import (
	"encoding/json"
	"sync"

	"github.com/biogo/store/llrb"
	"github.com/juju/errors"

	"github.com/dearcode/candy/util/log"
)

const (
	maxRegionEnd = 9973
)

var (
	//ErrRegionNotFound region not exist.
	ErrRegionNotFound = errors.New("region not found")
)

//Region 用户ID区间begin~end
type Region struct {
	Begin int
	End   int
	Host  string
}

//Compare for llrb.
func (r Region) Compare(c llrb.Comparable) int {
	b := c.(*Region)
	if b.Begin == r.Begin {
		return 0
	}

	if b.Begin < r.Begin && r.Begin < b.End {
		return 0
	}

	return b.Begin - r.Begin
}

func newRegion(host string, begin, end int) *Region {
	return &Region{Begin: begin, End: end, Host: host}
}

//Regions 保存region的路由信息.
type Regions struct {
	mu    sync.RWMutex
	tree  *llrb.Tree
	hosts map[string]*Region
}

//NewRegions 初始化llrb，map.
func NewRegions(ops ...RegionsOption) (*Regions, error) {
	r := &Regions{tree: &llrb.Tree{}, hosts: make(map[string]*Region)}
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

//span 区间的跨度
func (r *Region) span() int {
	return r.End - r.Begin
}

//Max 两个Region最大范围
func (r *Region) Max(c Region) (b int, e int) {
	b = r.Begin
	e = r.End

	if b > c.Begin {
		b = c.Begin
	}

	if e < c.End {
		e = c.End
	}

	return
}

//Max 当前跨度最大的节点
func (r *Regions) Max() (Region, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.tree.Len() == 0 {
		return Region{}, ErrRegionNotFound
	}

	var o *Region
	r.tree.Do(func(c llrb.Comparable) (done bool) {
		n := c.(*Region)
		if o == nil || n.span()-o.span() >= o.span()/2 {
			o = n
		}
		return
	})
	return *o, nil
}

//Match 用户与region匹配
func (r *Region) Match(id int64) bool {
	i := int(id % maxRegionEnd)
	if i >= r.Begin && i < r.End {
		return true
	}
	return false
}

//Split 添加一个节点并分割相对大的节点.
func (r *Regions) Split(from, to string) (Region, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tree.Len() == 0 {
		n := newRegion(to, 0, maxRegionEnd)
		r.tree.Insert(n)
		r.hosts[to] = n
		return *n, nil
	}

	f, ok := r.hosts[from]
	if !ok {
		log.Infof("from %s not found", from)
		return Region{}, ErrRegionNotFound
	}

	t, ok := r.hosts[to]
	if ok {
		log.Infof("new region %s exist", to)
		return *t, nil
	}

	delta := (f.End - f.Begin) / 2

	t = newRegion(to, f.End-delta, f.End)

	r.tree.Delete(f)
	f.End -= delta
	r.tree.Insert(f)

	r.tree.Insert(t)

	r.hosts[to] = t

	return *t, nil
}

//Get 根据ID获取对应的region信息
func (r *Regions) Get(id int64) (Region, bool) {
	b := int(id % maxRegionEnd)
	r.mu.RLock()
	c := r.tree.Get(&Region{Begin: b})
	if c == nil {
		return Region{}, false
	}
	src := c.(*Region)
	r.mu.RUnlock()

	return *src, true
}

//GetByHost 根据host获取对应的region信息.
func (r *Regions) GetByHost(host string) (Region, error) {
	r.mu.RLock()
	if src, ok := r.hosts[host]; ok {
		r.mu.RUnlock()
		return *src, ErrRegionNotFound
	}
	r.mu.RUnlock()

	return Region{}, nil
}

// Next 返回当前节点的下一个节点
func (r *Regions) Next(host string) (self Region, next Region, err error) {
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
	if begin == maxRegionEnd {
		begin = f.Begin - 1
	}

	c := r.tree.Get(&Region{Begin: begin})
	if c == nil {
		log.Infof("region begin:%d not found", begin)
		err = ErrRegionNotFound
		return
	}

	self = *f
	next = *c.(*Region)

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
func (r *Regions) Marshal() (string, error) {
	r.mu.RLock()
	buf, err := json.Marshal(r.hosts)
	r.mu.RUnlock()
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
