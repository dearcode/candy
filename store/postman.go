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

// send 分发消息
// 推送流程：
// 1.先添加到收件人消息列表中
// 2.再添加到发件人的消息列表中
// 3.调用notice接口Push通知
// 如果是群发，需要依次添加到每个人的消息列表中
func (p *postman) send(msg meta.Message) error {
	log.Debugf("msg:%v", msg)
	if err := p.check(msg); err != nil {
		return errors.Trace(err)
	}

	//私聊：只给一个发就退出了
	if msg.Group == 0 {
		if err := p.user.addMessage(msg.User, msg.ID); err != nil {
			return errors.Trace(err)
		}
		return p.notice.Push(msg, msg.User)
	}

	//群聊：向组中每个人添加消息
	group, err := p.group.get(msg.Group)
	if err != nil {
		return errors.Trace(err)
	}

	for _, uid := range group.Users {
		if err := p.user.addMessage(uid, msg.ID); err != nil {
			return errors.Trace(err)
		}
	}

	return p.notice.Push(msg, group.Users...)
}
