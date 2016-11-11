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
	users  map[int64]devices
	sender brokerSender
	sync.RWMutex
}

type brokerSender interface {
	push(addr string, req meta.PushRequest) error
}

func newBroker(sender brokerSender) *broker {
	return &broker{users: make(map[int64]devices), sender: sender}
}

// classify 按用户所在Gate地址拆分PushID
func (b *broker) classify(pids []meta.PushID) map[string][]meta.PushID {
	hosts := make(map[string][]meta.PushID)
	b.RLock()
	for _, id := range pids {
		devs, ok := b.users[id.User]
		if !ok {
			continue
		}

		for _, dev := range devs {
			if ids, ok := hosts[dev.host]; ok {
				ids = append(ids, id)
				continue
			}
			hosts[dev.host] = []meta.PushID{id}
		}
	}
	b.RUnlock()
	return hosts
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

// send 这里做拆分推送: 一个推送消息，可能推送到多个gate上用户的多个设备也可能用户不在线
func (b *broker) send(msg meta.PushMessage, pids []meta.PushID) []meta.PushID {
	log.Debugf("send msg:%v pids:%v", msg, pids)
	s := make(map[int64]struct{})

	for host, pi := range b.classify(pids) {
		req := meta.PushRequest{ID: pi, Msg: msg}
		if err := b.sender.push(host, req); err != nil {
			log.Errorf("push %s req:%+v, err:%s", host, req, errors.ErrorStack(err))
			continue
		}

		for _, id := range pi {
			s[id.User] = struct{}{}
		}
	}

	var eps []meta.PushID
	for _, pi := range pids {
		if _, ok := s[pi.User]; !ok {
			eps = append(eps, pi)
		}
	}

	return eps
}
