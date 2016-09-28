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
func (p *postman) check(msg meta.Message) error {
	log.Debugf("msg check")
	if msg.User == 0 {
		// 检测用户是否在组中
		if err := p.group.exist(msg.Group, msg.From); err != nil {
			return errors.Annotatef(ErrInvalidSender, "group:%d not found user:%d", msg.Group, msg.From)
		}
		return nil
	}

	// 检测用户是否为好友
	log.Debugf("whether is friend")
	if err := p.friend.exist(msg.From, msg.User); err != nil {
		return errors.Annotatef(ErrInvalidSender, "unrelated from:%d to:%d", msg.From, msg.User)
	}

	log.Debugf("success")
	return nil
}

// sendToUser 发给单个用户
func (p *postman) sendToUser(msg meta.Message) ([]*meta.PushID, error) {
	before, err := p.user.addMessage(msg.User, msg.ID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return []*meta.PushID{{Before: before, User: msg.User}}, nil
}

// sendToGroup 发送给群
func (p *postman) sendToGroup(msg meta.Message) ([]*meta.PushID, error) {
	var ids []*meta.PushID
	group, err := p.group.get(msg.Group)
	if err != nil {
		return nil, errors.Trace(err)
	}

	//向组中每个人添加消息
	for _, uid := range group.Users {
		before, err := p.user.addMessage(uid, msg.ID)
		if err != nil {
			return nil, errors.Trace(err)
		}
		ids = append(ids, &meta.PushID{Before: before, User: uid})
	}

	return ids, nil
}

// send 分发消息
// 1.检查是否存在好友关系或者群成员
// 2.添加到收件人消息列表中
// 3.调用notice接口Push通知
func (p *postman) send(msg meta.Message) error {
	log.Debugf("msg:%v", msg)
	err := p.check(msg)
	if err != nil {
		return errors.Trace(err)
	}

	var ids []*meta.PushID

	if msg.Group == 0 {
		ids, err = p.sendToUser(msg)
	} else {
		ids, err = p.sendToGroup(msg)
	}

	if err != nil {
		return errors.Trace(err)
	}

	return p.notice.Push(msg, ids...)
}
