package store

import (
	"github.com/juju/errors"

	"github.com/dearcode/candy/server/meta"
)

type postman struct {
	user   *userDB
	friend *friendDB
	group  *groupDB
}

func newPostman() *postman {
	return &postman{}
}

func (p *postman) check(msg meta.Message) error {
	if msg.User == 0 {
		// 检测用户是否在组中
		if err := p.group.exist(msg.Group, msg.From); err != nil {
			return errors.Trace(err)
		}
		return nil
	}

	// 检测用户是否为好友
	if err := p.friend.exist(msg.From, msg.User); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// 推送流程：
// 1.先添加到收件人消息列表中
// 2.再添加到发件人的消息列表中
// 3.调用notice发通知
// 如果是群发，需要依次添加到每个人的消息列表中
func (p *postman) push(msg meta.Message) error {
	if err := p.check(msg); err != nil {
		return errors.Trace(err)
	}

	if msg.Group == 0 {
		if err := p.user.addMessage(msg.User, msg.ID); err != nil {
			return errors.Trace(err)
		}
		return p.notice(msg.User, msg)
	}

	//向组中每个人添加消息
	group, err := p.group.get(msg.Group)
	if err != nil {
		return errors.Trace(err)
	}
	for _, uid := range group.Users {
		if err := p.user.addMessage(uid, msg.ID); err != nil {
			return errors.Trace(err)
		}
		p.notice(uid, msg)
	}

	return nil
}

// 调用notice发推送消息
func (p *postman) notice(uid int64, msg meta.Message) error {

	return nil
}
