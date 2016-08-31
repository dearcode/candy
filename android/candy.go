package candy

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

const (
	networkTimeout = time.Second * 3
)

type Candy struct {
	host string
	conn *grpc.ClientConn
	api  meta.GateClient
}

func NewCandy(host string) *Candy {
	return &Candy{host: host}
}

func (c *Candy) Start() (err error) {
	if c.conn, err = grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout)); err != nil {
		return
	}

	c.api = meta.NewGateClient(c.conn)
    return
}

func (c *Candy) Register(user, passwd string) (int64, error) {
	req := &meta.UserRegisterRequest{User: user, Password: passwd}
	resp, err := c.api.Register(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header
}

