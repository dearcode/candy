package gate

import (
	"net"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

var (
	ErrUndefineMethod = errors.New("undefine method.")
)

// gateServer recv client request.
type gateServer struct {
	host string
}

func NewGateServer(host string) *gateServer {
	return &gateServer{host: host}
}

func (g *gateServer) Start() error {
	serv := grpc.NewServer()
	meta.RegisterMessageServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return err
	}

	return serv.Serve(lis)
}

func (g *gateServer) Register(_ context.Context, req *meta.UserRegisterRequest) (*meta.UserRegisterResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) UpdateUserInfo(_ context.Context, req *meta.UpdateUserInfoRequest) (*meta.UpdateUserInfoResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) Login(_ context.Context, req *meta.UserLoginRequest) (*meta.UserLoginResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) Logout(_ context.Context, req *meta.UserLogoutRequest) (*meta.UserLoginResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) SendMessage(_ context.Context, req *meta.SendMessageRequest) (*meta.SendMessageResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) RecvMessage(_ context.Context, req *meta.RecvMessageRequest) (*meta.RecvMessageResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) Heartbeat(_ context.Context, req *meta.HeartbeatRequest) (*meta.HeartbeatResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) UploadImage(_ context.Context, req *meta.UploadImageRequest) (*meta.UploadImageResponse, error) {
	return nil, ErrUndefineMethod
}

func (g *gateServer) DownloadImage(_ context.Context, req *meta.DownloadImageRequest) (*meta.DownloadImageResponse, error) {
	return nil, ErrUndefineMethod
}
