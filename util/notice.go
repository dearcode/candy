package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// Notifer 连接Push服务.
type Notifer struct {
	conn   *grpc.ClientConn
	client meta.PushClient
	stream meta.Push_SubscribeClient
	push   chan<- meta.PushRequest
}

// NewNotifer 返回Notifer client.
func NewNotifer(host string) (*Notifer, error) {
	var err error

	n := &Notifer{}
	log.Debugf("dial host:%v", host)
	if n.conn, err = grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout)); err != nil {
		return n, errors.Trace(err)
	}

	n.client = meta.NewPushClient(n.conn)
	if n.stream, err = n.client.Subscribe(context.Background()); err != nil {
		n.conn.Close()
		return n, errors.Trace(err)
	}

	return n, nil
}

// Recv 接收stream消息
func (n *Notifer) Recv() (*meta.PushRequest, error) {
	return n.stream.Recv()
}

//Subscribe 订阅消息
func (n *Notifer) Subscribe(id int64, device string) error {
	req := &meta.SubscribeRequest{ID: id, Enable: true, Device: device}
	return errors.Trace(n.stream.Send(req))
}

//UnSubscribe 取消订阅消息, 取消哪个渠道的推送
func (n *Notifer) UnSubscribe(id int64, device string) error {
	req := &meta.SubscribeRequest{ID: id, Enable: false, Device: device}
	return errors.Trace(n.stream.Send(req))
}

//Push  调用Notifer发推送消息
func (n *Notifer) Push(msg meta.PushMessage, ids ...meta.PushID) error {
	req := &meta.PushRequest{ID: ids, Msg: msg}
	resp, err := n.client.Push(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}
