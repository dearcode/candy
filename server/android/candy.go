package candy

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
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

	return userInfo, resp.Header.Error()
}

func (c *CandyClient) AddFriend(userID int64, confirm bool) (bool, error) {
	req := &meta.GateAddFriendRequest{UserID: userID, Confirm: confirm}
	resp, err := c.api.AddFriend(context.Background(), req)
	if err != nil {
		return false, err
	}

	return resp.Confirm, resp.Header.Error()
}

type UserList struct {
	Users []*UserInfo
}

// 支持模糊查询，返回对应用户的列表
func (c *CandyClient) FindUser(user string) (*UserList, error) {
	req := &meta.GateFindUserRequest{User: user}
	resp, err := c.api.FindUser(context.Background(), req)
	if err != nil {
		return nil, err
	}

	users := make([]*UserInfo, 0)
	for _, matchUser := range resp.Users {
		userInfo, err := c.GetUserInfo(matchUser)
		if err != nil {
			return nil, err
		}
		users = append(users, userInfo)
	}

	return &UserList{Users: users}, resp.Header.Error()
}

func (c *CandyClient) FileExist(key string) (bool, error) {
	req := &meta.GateCheckFileRequest{Names: []string{key}}
	resp, err := c.api.CheckFile(context.Background(), req)
	if err != nil {
		return false, err
	}

	if err = resp.Header.Error(); err != nil {
		return false, err
	}

	if len(resp.Names) == 0 {
		return true, nil
	}

	return false, nil
}

func (c *CandyClient) FileUpload(data []byte) (string, error) {
	md5 := string(util.MD5(data))
	exist, err := c.FileExist(md5)
	if err != nil {
		return md5, err
	}
	//已有别人上传过了
	if exist {
		return md5, nil
	}

	req := &meta.GateUploadFileRequest{File: data}
	resp, err := c.api.UploadFile(context.Background(), req)
	if err != nil {
		return md5, err
	}

	return md5, resp.Header.Error()
}

func (c *CandyClient) FileDownload(key string) ([]byte, error) {
	req := &meta.GateDownloadFileRequest{Names: []string{key}}
	resp, err := c.api.DownloadFile(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Files[key], resp.Header.Error()
}

func (c *CandyClient) NewMessage(from, group, user int64, body string) error {
	gateClient, err := c.api.NewMessage(context.Background())
	if err != nil {
		return err
	}

	newMsg := &meta.Message{From: from, Group: group, User: user, Body: body}
	err = gateClient.Send(newMsg)
	if err != nil {
		return err
	}

	go c.recvMsg(gateClient)

	return nil
}

func (c *CandyClient) recvMsg(client meta.Gate_NewMessageClient) {
	msg, err := client.Recv()
	if err != nil {
		fmt.Println("接收消息失败， err:", err)
	}

	fmt.Println("接收消息成功, msg:", msg)

}
