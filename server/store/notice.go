package store

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
)

type notice struct {
	host   string
	notice meta.NoticeClient
}

func newNotice(host string) *notice {
	return &notice{host: host}
}

func (n *notice) start() error {
	log.Debugf("dial host:%v", n.host)
	conn, err := grpc.Dial(n.host, grpc.WithInsecure(), grpc.WithTimeout(util.NetworkTimeout))
	if err != nil {
		return errors.Trace(err)
	}

	n.notice = meta.NewNoticeClient(conn)
	if n.notice == nil {
		return errors.Errorf("create new notice client error, host:%v", n.host)
	}

	return nil
}

//调用notice订阅消息
func (n *notice) subscribe(id int64, host string) error {
	req := &meta.SubscribeRequest{ID: id, Host: host}
	if n.notice == nil {
		return errors.New("error notice is nil")
	}

	resp, err := n.notice.Subscribe(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}

//调用notice取消订阅消息
func (n *notice) unSubscribe(id int64, host string) error {
	req := &meta.UnSubscribeRequest{ID: id, Host: host}
	if n.notice == nil {
		return errors.New("error notice is nil")
	}

	resp, err := n.notice.UnSubscribe(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	if resp.Header != nil {
		return errors.New(resp.Header.Msg)
	}

	return nil
}
