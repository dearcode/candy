package gate

import (
	"net"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

const (
	networkTimeout = time.Second * 3
)

var (
	// ErrUndefineMethod 方法未定义
	ErrUndefineMethod = errors.New("undefine method")
)

// Gate recv client request.
type Gate struct {
	host   string
	master *master
	store  *store
}

// NewGate new gate server.
func NewGate(host, master, store string) *Gate {
	return &Gate{host: host, master: newMaster(master), store: newStore(store)}
}

// Start start service.
func (g *Gate) Start() error {
	serv := grpc.NewServer()
	meta.RegisterGateServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return errors.Trace(err)
	}

	if err = g.master.start(); err != nil {
		return errors.Trace(err)
	}

	if err = g.store.start(); err != nil {
		return errors.Trace(err)
	}

	return serv.Serve(lis)
}

// Register user, passwd.
func (g *Gate) Register(_ context.Context, req *meta.UserRegisterRequest) (*meta.UserRegisterResponse, error) {
	id, err := g.master.newID()
	if err != nil {
		return &meta.UserRegisterResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	if err = g.store.register(req.User, req.Password, id); err != nil {
		return &meta.UserRegisterResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.UserRegisterResponse{ID: id}, nil
}

// UpdateUserInfo nickname.
func (g *Gate) UpdateUserInfo(_ context.Context, req *meta.UpdateUserInfoRequest) (*meta.UpdateUserInfoResponse, error) {
	return nil, ErrUndefineMethod
}

// Login user,passwd.
func (g *Gate) Login(_ context.Context, req *meta.UserLoginRequest) (*meta.UserLoginResponse, error) {
	id, err := g.store.auth(req.User, req.Password)
	if err != nil {
		return &meta.UserLoginResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.UserLoginResponse{ID: id}, nil
}

// Logout nil.
func (g *Gate) Logout(_ context.Context, req *meta.UserLogoutRequest) (*meta.UserLoginResponse, error) {
	return nil, ErrUndefineMethod
}

// SendMessage from,to,msg.
func (g *Gate) SendMessage(_ context.Context, req *meta.SendMessageRequest) (*meta.SendMessageResponse, error) {
	return nil, ErrUndefineMethod
}

// RecvMessage nil.
func (g *Gate) RecvMessage(_ context.Context, req *meta.RecvMessageRequest) (*meta.RecvMessageResponse, error) {
	return nil, ErrUndefineMethod
}

// Heartbeat nil.
func (g *Gate) Heartbeat(_ context.Context, req *meta.HeartbeatRequest) (*meta.HeartbeatResponse, error) {
	return nil, ErrUndefineMethod
}

// UploadImage image.
func (g *Gate) UploadImage(_ context.Context, req *meta.UploadImageRequest) (*meta.UploadImageResponse, error) {
	return nil, ErrUndefineMethod
}

// DownloadImage ids.
func (g *Gate) DownloadImage(_ context.Context, req *meta.DownloadImageRequest) (*meta.DownloadImageResponse, error) {
	return nil, ErrUndefineMethod
}

// Notice recv Notice server Message, and send Message to client.
func (g *Gate) Notice(_ context.Context, req *meta.NoticeRequest) (*meta.NoticeResponse, error) {
	return nil, ErrUndefineMethod
}
