package candy

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
)

const (
	networkTimeout = time.Second * 3
)

type CandyClient struct {
	host string
	conn *grpc.ClientConn
	api  meta.GateClient
}

func NewCandyClient(host string) *CandyClient {
	return &CandyClient{host: host}
}

func (c *CandyClient) Start() (err error) {
	if c.conn, err = grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithTimeout(networkTimeout)); err != nil {
		return
	}

	c.api = meta.NewGateClient(c.conn)
	return
}

func (c *CandyClient) Register(user, passwd string) (int64, error) {
	req := &meta.GateRegisterRequest{User: user, Password: passwd}
	resp, err := c.api.Register(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

func (c *CandyClient) Login(user, passwd string) (int64, error) {
	req := &meta.GateUserLoginRequest{User: user, Password: passwd}
	resp, err := c.api.Login(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

func (c *CandyClient) Logout(user string) error {
	req := &meta.GateUserLogoutRequest{User: user}
	resp, err := c.api.Logout(context.Background(), req)
	if err != nil {
		return err
	}

	return resp.Header.Error()
}

func (c *CandyClient) UpdateUserInfo(user, nickName string, avatar []byte) (int64, error) {
	req := &meta.GateUpdateUserInfoRequest{User: user, NickName: nickName, Avatar: avatar}
	resp, err := c.api.UpdateUserInfo(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

func (c *CandyClient) UpdateUserPassword(user, passwd string) (int64, error) {
	req := &meta.GateUpdateUserPasswordRequest{User: user, Password: passwd}
	resp, err := c.api.UpdateUserPassword(context.Background(), req)
	if err != nil {
		return -1, err
	}

	return resp.ID, resp.Header.Error()
}

func (c *CandyClient) GetUserInfo(user string) (*UserInfo, error) {
	req := &meta.GateGetUserInfoRequest{User: user}
	resp, err := c.api.GetUserInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}

	userInfo := &UserInfo{ID: resp.ID, Name: resp.User, NickName: resp.NickName, Avatar: resp.Avatar}

	return userInfo, nil
}
