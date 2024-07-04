package server

import (
	"context"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type MasterServer struct {
	BaseServer
}

func (s *MasterServer) NodeType() string {
	return "master"
}

func NewMasterServer(logger log.Logger, opts ServerOptions) (MasterServer, error) {
	baseServer, err := NewBaseServer(logger, opts)
	if err != nil {
		return MasterServer{}, fmt.Errorf("error initializing master server: %w", err)
	}
	return MasterServer{
		BaseServer: baseServer,
	}, nil
}

func (s *MasterServer) Run(ctx context.Context) error {
	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		s,
	)
	go s.ConnectionHandler(ctx)
	go s.ExpiryLoop(ctx)

	return nil
}
