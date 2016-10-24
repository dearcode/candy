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
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout))
	if err != nil {
		return nil, errors.Trace(err)
	}

	client := meta.NewPushClient(conn)

	stream, err := client.Subscribe(context.Background())
	if err != nil {
		conn.Close()
		return nil, errors.Trace(err)
	}

	return &Notifer{conn: conn, stream: stream, client: client}, nil
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
