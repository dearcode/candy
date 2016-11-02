package notice

import (
	"sync"

	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type pushRequest struct {
	ids []meta.PushID
	msg meta.PushMessage
}

type broker struct {
	mbox   chan pushRequest
	users  map[int64]devices
	sender brokerSender
	sync.RWMutex
}

const (
	defaultPushChanSize = 1000
	// 默认推送线程数量
	defaultSenderNumber = 4
)

type brokerSender interface {
	push(addr string, req meta.PushRequest) error
}

func newBroker(sender brokerSender) *broker {
	b := &broker{users: make(map[int64]devices), mbox: make(chan pushRequest, defaultPushChanSize), sender: sender}
	for i := 0; i < defaultSenderNumber; i++ {
		go b.loopSender()
	}
	return b
}

// split 按用户所在chan拆分请求
func (b *broker) split(req pushRequest) map[string][]meta.PushID {
	pids := make(map[string][]meta.PushID)
	b.RLock()
	for _, id := range req.ids {
		devs, ok := b.users[id.User]
		if !ok {
			// 跳过未订阅的用户
			continue
		}

		for _, dev := range devs {
			if ids, ok := pids[dev.host]; ok {
				ids = append(ids, id)
				continue
			}
			pids[dev.host] = []meta.PushID{id}
		}
	}
	b.RUnlock()
	return pids
}

// loopSender 这里做拆分推送: 一个推送消息，可能推送到多个gate上用户的多个设备也可能用户不在线
func (b *broker) loopSender() {
	for {
		req := <-b.mbox
		for host, ids := range b.split(req) {
			req := meta.PushRequest{ID: ids, Msg: req.msg}
			if err := b.sender.push(host, req); err != nil {
				log.Errorf("push %s req:%+v, err:%s", host, req, errors.ErrorStack(err))
			}
		}
	}
}

func (b *broker) subscribe(uid, token int64, dev, host string) {
	b.Lock()
	defer b.Unlock()

	devs, ok := b.users[uid]
	if !ok {
		devs = newDevices()
		b.users[uid] = devs
	}

	devs.put(dev, device{token: token, host: host})

	log.Debugf("subscribe uid:%d, dev:%s, host:%s, token:%d", uid, dev, host, token)
}

func (b *broker) unSubscribe(uid, token int64, dev string) {
	b.Lock()
	if devs, ok := b.users[uid]; ok {
		devs.del(dev, token)
		if devs.empty() {
			delete(b.users, uid)
		}
	}
	b.Unlock()
	log.Debugf("unSubscribe uid:%d, dev:%s, token:%d", uid, dev, token)
}

func (b *broker) push(msg meta.PushMessage, ids ...meta.PushID) {
	log.Debugf("broker msg:%v ids:%v", msg, ids)
	b.mbox <- pushRequest{ids: ids, msg: msg}
}
