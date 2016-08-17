package store

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dearcode/candy/meta"
)

// Store save user, message.
type Store struct {
	host   string
	dbPath string
	user   *userDB
}

// NewStore new Store server.
func NewStore(host, dbPath string) *Store {
	return &Store{host: host, dbPath: dbPath, user: newUserDB(dbPath)}
}

// Start start service.
func (s *Store) Start() error {
	serv := grpc.NewServer()
	meta.RegisterStoreServer(serv, s)

	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	if err = s.user.start(); err != nil {
		return err
	}

	return serv.Serve(lis)
}

// Register add user.
func (s *Store) Register(_ context.Context, req *meta.RegisterRequest) (*meta.RegisterResponse, error) {
	if err := s.user.register(req.User, req.Password, req.ID); err != nil {
		return &meta.RegisterResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}

	return &meta.RegisterResponse{}, nil
}

// Auth check password.
func (s *Store) Auth(_ context.Context, req *meta.AuthRequest) (*meta.AuthResponse, error) {
	id, err := s.user.auth(req.User, req.Password)
	if err != nil {
		return &meta.AuthResponse{Header: &meta.ResponseHeader{Code: -1, Msg: err.Error()}}, nil
	}
	return &meta.AuthResponse{ID: id}, nil
}
