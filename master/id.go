package master

import (
	"sync"
	"time"
)

type idAllocator struct {
	lastID int64
	seq    uint32
	sync.Mutex
}

// 当前阶段只在本地记录，不考虑分布式，以后转移到etcd中
func newIDAllocator() *idAllocator {
	return &idAllocator{}
}

func (i *idAllocator) newID() int64 {
	i.Lock()
	defer i.Unlock()

	c := time.Now().Unix()
	if i.lastID != c {
		i.seq = 0
		i.lastID = c
	} else {
		i.seq++
	}

	return int64(i.lastID<<32) + int64(i.seq)
}
