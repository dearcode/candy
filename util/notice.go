package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// NotiferClient 连接Notifer服务.
type NotiferClient struct {
	client meta.NotiferClient
}

// NewNotiferClient 返回NotiferClient client.
func NewNotiferClient(host string) (*NotiferClient, error) {
	log.Debugf("dial host:%v", host)
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &NotiferClient{client: meta.NewNotiferClient(conn)}, nil
}

//Subscribe 为gate提供，调用Notifer订阅消息
func (n *NotiferClient) Subscribe(id int64, device string, token int64, host string) error {
	req := &meta.SubscribeRequest{ID: id, Device: device, Host: host, Token: token}
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	resp, err := n.client.Subscribe(ctx, req)
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}

//UnSubscribe 为gate提供，调用Notifer取消订阅消息
func (n *NotiferClient) UnSubscribe(id int64, device, host string, sid int64) error {
	req := &meta.UnSubscribeRequest{ID: id, Device: device, Host: host, Token: sid}
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	resp, err := n.client.UnSubscribe(ctx, req)
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}

//Push 为store提供调用Notifer发推送消息
func (n *NotiferClient) Push(msg meta.PushMessage, ids ...meta.PushID) error {
	req := &meta.PushRequest{ID: ids, Msg: msg}
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	resp, err := n.client.Push(ctx, req)
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}
