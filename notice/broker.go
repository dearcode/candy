package notice

import (
	"sync"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type pushRequest struct {
	ids []meta.PushID
	msg meta.PushMessage
}

type device struct {
	//cid 所在chan的ID
	cid  int32
	name string
}

type broker struct {
	mbox      chan pushRequest
	chans     map[int32]chan pushRequest
	chanIndex int32
	chansLock sync.RWMutex

	// users 存储用户ID对应的chan列表
	users map[int64][]device
	sync.RWMutex
}

const (
	defaultPushChanSize = 1000
	// 默认推送线程数量
	defaultSenderNumber = 4
)

func newBroker() *broker {
	return &broker{users: make(map[int64][]device), chans: make(map[int32]chan pushRequest), mbox: make(chan pushRequest, defaultPushChanSize)}
}

// Start start service.
func (b *broker) start() {
	go b.run()
}

// split 按用户所在chan拆分请求
func (b *broker) split(req pushRequest) map[int32][]meta.PushID {
	pids := make(map[int32][]meta.PushID)
	b.RLock()
	for _, id := range req.ids {
		devs, ok := b.users[id.User]
		if !ok {
			// 跳过未订阅的用户
			continue
		}

		for _, dev := range devs {
			if ids, ok := pids[dev.cid]; ok {
				ids = append(ids, id)
				continue
			}
			pids[dev.cid] = []meta.PushID{id}
		}
	}
	b.RUnlock()
	return pids
}

// run 这里做拆分推送: 一个推送消息，可能推送到多个gate上用户的多个设备也可能用户不在线
func (b *broker) run() {
	for {
		req := <-b.mbox
		for cid, ids := range b.split(req) {
			req := pushRequest{ids: ids, msg: req.msg}
			b.chansLock.RLock()
			if c, ok := b.chans[cid]; ok {
				c <- req
			}
			b.chansLock.RUnlock()
		}
	}
}

//addPushChan gate的连接上来要加，断开要减去
func (b *broker) addPushChan(c chan pushRequest) int32 {
	var cid int32
	b.chansLock.Lock()
	b.chanIndex++
	cid = b.chanIndex
	b.chans[cid] = c
	b.chansLock.Unlock()
	return cid
}

//delPushChan gate的连接上来要加，断开要减去
func (b *broker) delPushChan(cid int32) {
	b.chansLock.Lock()
	delete(b.chans, cid)
	b.chansLock.Unlock()
}

func (b *broker) subscribe(uid int64, dev string, cid int32) {
	log.Debugf("Subscribe uid:%d, dev:%s, cid:%d", uid, dev, cid)
	b.Lock()
	if devs, ok := b.users[uid]; ok {
		for i := 0; i < len(devs); {
			log.Debugf("dev[%d]:%s, devname:%s", i, devs[i].name, dev)
			if devs[i].name == dev {
				copy(devs[i:], devs[i+1:])
				devs = devs[:len(devs)-1]
				log.Debugf("new devs:%v", devs)
				continue
			}
			i++
		}
		b.users[uid] = devs
	}
	b.users[uid] = append(b.users[uid], device{name: dev, cid: cid})
	b.Unlock()
}

func (b *broker) unSubscribe(uid int64, dev string) {
	log.Debugf("UnSubscribe uid:%d, dev:%s", uid, dev)
	b.Lock()
	if devs, ok := b.users[uid]; ok {
		for i := 0; i < len(devs); {
			if devs[i].name == dev {
				copy(devs[i:], devs[i+1:])
				devs = devs[:len(devs)-1]
			}
		}

		if len(devs) == 0 {
			delete(b.users, uid)
		}
	}
	b.Unlock()
}

func (b *broker) push(msg meta.PushMessage, ids ...meta.PushID) {
	log.Debugf("broker msg:%v ids:%v", msg, ids)
	b.mbox <- pushRequest{ids: ids, msg: msg}
}
