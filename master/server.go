package master

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

// Master process gate request.
type Master struct {
	host        string
	idAllocator *idAllocator
}

// NewMaster create new Master.
func NewMaster(host string) *Master {
	return &Master{host: host, idAllocator: newIDAllocator()}
}

// Start start Master
func (g *Master) Start() error {
	log.Debugf("Master Start...")
	serv := grpc.NewServer()
	meta.RegisterMasterServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return err
	}

	return serv.Serve(lis)
}

// NewID return an new id
func (g *Master) NewID(_ context.Context, _ *meta.NewIDRequest) (*meta.NewIDResponse, error) {
	return &meta.NewIDResponse{ID: g.idAllocator.newID()}, nil
}
