package store

import (
	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type postman struct {
	user   *userDB
	friend *friendDB
	group  *groupDB
	notice *util.Notice
}

var (
	// ErrInvalidSender invalid sender.
	ErrInvalidSender = errors.New("invalid sender")
)

func newPostman(user *userDB, friend *friendDB, group *groupDB, notice *util.Notice) *postman {
	return &postman{user: user, friend: friend, group: group, notice: notice}
}

// check 检查发件人是否有权发送这个消息.
func (p *postman) check(msg *meta.Message) error {
	if msg.Group != 0 {
		// 检测组中是否存在这个发件人
		if err := p.group.exist(msg.Group, msg.From); err != nil {
			return errors.Annotatef(ErrInvalidSender, "group:%d not found user:%d", msg.Group, msg.From)
		}
		return nil
	}

	// 检测收件人的好友里面有没有发件人
	if err := p.friend.exist(msg.To, msg.From); err != nil {
		return errors.Annotatef(ErrInvalidSender, "unrelated from:%d to:%d", msg.From, msg.To)
	}

	return nil
}

// sendToUser 发给单个用户
func (p *postman) sendToUser(pm meta.PushMessage) error {
	before, err := p.user.addMessage(pm.Msg.To, pm.Msg.ID)
	if err != nil {
		return errors.Trace(err)
	}
	log.Debugf("send msg:%v, ids:%v", pm, pm.Msg.To)
	return p.notice.Push(pm, &meta.PushID{Before: before, User: pm.Msg.To})
}

// sendToGroup 发送给群
func (p *postman) sendToGroup(pm meta.PushMessage) error {
	var ids []*meta.PushID

	group, err := p.group.get(pm.Msg.Group)
	if err != nil {
		return errors.Trace(err)
	}

	//向组中每个人添加消息
	for _, uid := range group.Users {
		before, err := p.user.addMessage(uid, pm.Msg.ID)
		if err != nil {
			return errors.Trace(err)
		}
		ids = append(ids, &meta.PushID{Before: before, User: uid})
	}

	log.Debugf("send msg:%v, ids:%v", pm, ids)
	return p.notice.Push(pm, ids...)
}

// send 分发消息
// 1.检查是否存在好友关系或者群成员
// 2.添加到收件人消息列表中
// 3.调用notice接口Push通知
func (p *postman) send(pm meta.PushMessage) error {
	log.Debugf("begin send msg:%v", pm)

	if pm.Event == meta.Event_None {
		if err := p.check(pm.Msg); err != nil {
			log.Debugf("check error:%s", errors.ErrorStack(err))
			return errors.Trace(err)
		}
	}

	if pm.Msg.Group == 0 {
		return p.sendToUser(pm)
	}

	return p.sendToGroup(pm)

}
