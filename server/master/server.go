package master

import (
	"net"

	"github.com/ngaut/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
)

// masterServer process gate request.
type masterServer struct {
	host        string
	idAllocator *idAllocator
}

func NewMasterServer(host string) *masterServer {
	return &masterServer{host: host, idAllocator: newIDAllocator()}
}

func (g *masterServer) Start() error {
	log.Debugf("masterServer Start...")
	serv := grpc.NewServer()
	meta.RegisterMasterServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return err
	}

	return serv.Serve(lis)
}

func (g *masterServer) NewID(_ context.Context, _ *meta.NewIDRequest) (*meta.NewIDResponse, error) {
	return &meta.NewIDResponse{ID: g.idAllocator.newID()}, nil
}
