package store

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
)

type postman struct {
	host   string
	user   *userDB
	friend *friendDB
	group  *groupDB
	notice meta.NoticeClient
}

func newPostman(host string, user *userDB, friend *friendDB, group *groupDB) *postman {
	return &postman{host: host, user: user, friend: friend, group: group}
}

func (p *postman) start() error {
	log.Debugf("dial host:%v", p.host)
	conn, err := grpc.Dial(p.host, grpc.WithInsecure(), grpc.WithTimeout(util.NetworkTimeout))
	if err != nil {
		return errors.Trace(err)
	}

	p.notice = meta.NewNoticeClient(conn)
	if p.notice == nil {
		return errors.Errorf("create new notice client error, host:%v", p.host)
	}

	return nil
}

func (p *postman) check(msg meta.Message) error {
	log.Debugf("msg check")
	if msg.User == 0 {
		// 检测用户是否在组中
		if err := p.group.exist(msg.Group, msg.From); err != nil {
			return errors.Trace(err)
		}
		return nil
	}

	// 检测用户是否为好友
	log.Debugf("whether is friend")
	/*if err := p.friend.exist(msg.From, msg.User); err != nil {
		return errors.Trace(err)
	}
	*/

	log.Debugf("success")
	return nil
}

// 推送流程：
// 1.先添加到收件人消息列表中
// 2.再添加到发件人的消息列表中
// 3.调用notice发通知
// 如果是群发，需要依次添加到每个人的消息列表中
func (p *postman) send(msg meta.Message) error {
	log.Debugf("msg:%v", msg)
	if err := p.check(msg); err != nil {
		return errors.Trace(err)
	}

	//私聊
	if msg.Group == 0 {
		if err := p.user.addMessage(msg.User, msg.ID); err != nil {
			return errors.Trace(err)
		}
		return p.push(msg, msg.User)
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
	}

	p.push(msg, group.Users...)

	return nil
}

// 调用notice发推送消息
func (p *postman) push(msg meta.Message, ids ...int64) error {
	log.Debugf("msg:%v ids:%v", msg, ids)
	req := &meta.PushRequest{ID: ids, Msg: &msg}
	if p.notice == nil {
		return errors.New("error notice is nil")
	}

	resp, err := p.notice.Push(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}
