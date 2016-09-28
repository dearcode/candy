package master

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// MasterServer process gate request.
type MasterServer struct {
	host        string
	idAllocator *idAllocator
}

// NewMasterServer create new MasterServer
func NewMasterServer(host string) *MasterServer {
	return &MasterServer{host: host, idAllocator: newIDAllocator()}
}

// Start start masterServer
func (g *MasterServer) Start() error {
	log.Debugf("masterServer Start...")
	serv := grpc.NewServer()
	meta.RegisterMasterServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return err
	}

	return serv.Serve(lis)
}

// NewID return an new id
func (g *MasterServer) NewID(_ context.Context, _ *meta.NewIDRequest) (*meta.NewIDResponse, error) {
	return &meta.NewIDResponse{ID: g.idAllocator.newID()}, nil
}
