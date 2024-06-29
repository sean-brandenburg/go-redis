package server

import (
	"context"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"go.uber.org/zap"
)

type SlaveServer struct {
	BaseServer

	// The `hostname:port` string that should be used to connect
	// to the master of this slave's replica set
	masterAddress string
}

func NewSlaveServer(logger log.Logger, masterAddress string, opts ServerOptions) (SlaveServer, error) {
	baseServer, err := NewBaseServer(logger, opts)
	if err != nil {
		return SlaveServer{}, fmt.Errorf("error initializing slave server: %w", err)
	}

	return SlaveServer{
		BaseServer:    baseServer,
		masterAddress: masterAddress,
	}, nil
}

func (s *SlaveServer) NodeType() string {
	return "slave"
}

func (s *SlaveServer) Run(ctx context.Context) error {
	conn, err := net.Dial("tcp", s.masterAddress)
	if err != nil {
		return fmt.Errorf("failed to dial master address %q: %w", s.masterAddress, err)
	}

	encodedPing, err := command.Encode([]any{"PING"})
	if err != nil {
		return fmt.Errorf("failed to encode PING to master: %w", err)
	}

	_, err = conn.Write([]byte(encodedPing))
	if err != nil {
		return fmt.Errorf("failed to write PING to master at address %q: %w", s.masterAddress, err)
	}

	rawRes := make([]byte, 512)
	bytesRead, err := conn.Read(rawRes)
	if err != nil {
		return fmt.Errorf("failed to read PING response from master at address %q: %w", s.masterAddress, err)
	}
	strRes := string(rawRes[:bytesRead])
	// TODO: Might be worth adding a command type that we can parse this into for PONG
	if strRes != "+PONG\r\n" {
		return fmt.Errorf("received unexpected response from master node: %q", strRes)
	}

	s.logger.Info("successfully PINGed master node", zap.String("masterAddress", s.masterAddress))

	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		func(cmd command.Command) (string, error) {
			return executeCommand(s, cmd)
		},
	)
	go s.ConnectionHandler(ctx)
	go s.ExpiryLoop(ctx)

	return nil
}
