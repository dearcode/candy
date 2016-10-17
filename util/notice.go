package util

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// Notice 连接notice服务.
type Notice struct {
	conn   *grpc.ClientConn
	client meta.PushClient
	stream meta.Push_SubscribeClient
	push   chan<- *meta.PushRequest
}

// NewNotice 返回notice client.
func NewNotice(host string, push chan<- *meta.PushRequest) (*Notice, error) {
	var err error

	n := &Notice{}
	log.Debugf("dial host:%v", host)
	if n.conn, err = grpc.Dial(host, grpc.WithInsecure(), grpc.WithTimeout(NetworkTimeout)); err != nil {
		return n, errors.Trace(err)
	}

	n.client = meta.NewPushClient(n.conn)
	if n.stream, err = n.client.Subscribe(context.Background()); err != nil {
		n.conn.Close()
		return n, errors.Trace(err)
	}

	go n.run()

	return n, nil
}

//Subscribe 订阅消息
func (n *Notice) Subscribe(id int64, device string) error {
	req := &meta.SubscribeRequest{ID: id, Enable: true, Device: device}
	return errors.Trace(n.stream.Send(req))
}

//UnSubscribe 取消订阅消息, 取消哪个渠道的推送
func (n *Notice) UnSubscribe(id int64, device string) error {
	req := &meta.SubscribeRequest{ID: id, Enable: false, Device: device}
	return errors.Trace(n.stream.Send(req))
}

func (n *Notice) run() error {
	for {
		pr, err := n.stream.Recv()
		if err != nil {
			log.Errorf("stream Recv error:%s", errors.ErrorStack(err))
			time.Sleep(time.Second)
			continue
		}

		n.push <- pr
	}
}

//Push  调用notice发推送消息
func (n *Notice) Push(msg meta.PushMessage, ids ...*meta.PushID) error {
	req := &meta.PushRequest{ID: ids, Msg: &msg}
	resp, err := n.client.Push(context.Background(), req)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("resp:%v", resp)
	return errors.Trace(resp.Header.Error())
}
