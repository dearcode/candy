package session

import (
	"net"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/server/meta"
)

// Session recv client request.
type Session struct {
	host string
}

// NewSession new Session server.
func NewSession(host string) *Session {
	return &Session{host: host}
}

// Start start service.
func (g *Session) Start() error {
	serv := grpc.NewServer()
	meta.RegisterMessageServer(serv, g)

	lis, err := net.Listen("tcp", g.host)
	if err != nil {
		return err
	}

	return serv.Serve(lis)
}
