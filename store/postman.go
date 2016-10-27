package store

import (
	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type postman struct {
	user    *userDB
	friend  *friendDB
	group   *groupDB
	notifer *util.NotiferClient
}

var (
	// ErrInvalidSender invalid sender.
	ErrInvalidSender = errors.New("invalid sender")
)

func newPostman(user *userDB, friend *friendDB, group *groupDB, notifer *util.NotiferClient) *postman {
	return &postman{user: user, friend: friend, group: group, notifer: notifer}
}

func (p *postman) sendToUser(pm meta.PushMessage) error {
	log.Debugf("begin send to User:%d msg:%+v", pm.Msg.To, pm)

	if pm.Event == meta.Event_None {
		// 检测收件人的好友里面有没有发件人
		if err := p.friend.exist(pm.Msg.To, pm.Msg.From); err != nil {
			return errors.Annotatef(ErrInvalidSender, "unrelated from:%d to:%d", pm.Msg.From, pm.Msg.To)
		}
	}

	before, err := p.user.addMessage(pm.Msg.To, pm.Msg.ID)
	if err != nil {
		return errors.Trace(err)
	}

	id := meta.PushID{Before: before, User: pm.Msg.To}
	log.Debugf("send msg:%+v, id:%v", pm, id)
	return p.notifer.Push(pm, id)
}

func (p *postman) sendToGroup(pm meta.PushMessage) error {
	log.Debugf("begin send Group msg:%+v", pm)

	if pm.Event == meta.Event_None {
		// 检测组中是否存在这个发件人
		if err := p.group.exist(pm.Msg.Group, pm.Msg.From); err != nil {
			return errors.Annotatef(ErrInvalidSender, "group:%d not found user:%d", pm.Msg.Group, pm.Msg.From)
		}
	}

	var ids []meta.PushID

	group, err := p.group.getGroup(pm.Msg.Group)
	if err != nil {
		return errors.Annotatef(ErrInvalidSender, "group:%d not found from:%d to:%d", pm.Msg.Group, pm.Msg.From, pm.Msg.To)
	}

	//向组中每个人添加消息
	for _, uid := range group.Member {
		before, err := p.user.addMessage(uid, pm.Msg.ID)
		if err != nil {
			return errors.Trace(err)
		}
		ids = append(ids, meta.PushID{Before: before, User: uid})
	}

	log.Debugf("send to group, msg:%+v, ids:%v", pm, ids)
	return p.notifer.Push(pm, ids...)
}

func (p *postman) send(pm meta.PushMessage) error {
	if pm.ToUser {
		return p.sendToUser(pm)
	}
	return p.sendToGroup(pm)
}
