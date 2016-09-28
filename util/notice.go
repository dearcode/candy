package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type Notice struct {
	client meta.NoticeClient
}

func NewNotice(host string) (*Notice, error) {
	log.Debugf("dial host:%v", host)
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Notice{client: meta.NewNoticeClient(conn)}, nil
}

//Subscribe 调用notice订阅消息
func (n *Notice) Subscribe(id int64, host string) error {
	req := &meta.SubscribeRequest{ID: id, Host: host}
	resp, err := n.client.Subscribe(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}

//UnSubscribe 调用notice取消订阅消息
func (n *Notice) UnSubscribe(id int64, host string) error {
	req := &meta.UnSubscribeRequest{ID: id, Host: host}
	resp, err := n.client.UnSubscribe(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}

//Push  调用notice发推送消息
func (n *Notice) Push(msg meta.Message, ids ...*meta.PushID) error {
	req := &meta.PushRequest{ID: ids, Msg: &msg}
	resp, err := n.client.Push(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}
