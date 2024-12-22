package server

import (
	"context"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/connection"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type MasterServer struct {
	BaseServer

	// A list of replica connections that are currently registered with this master
	registeredReplicaConns []connection.Connection
}

func (s *MasterServer) NodeType() NodeType {
	return MasterNodeType
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

func (s *MasterServer) ExecuteCommand(conn connection.Connection, cmd command.Command) error {
	err := RunCommand(s, conn, cmd)
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	err = s.handleCommandPropagation(cmd)
	if err != nil {
		return fmt.Errorf("error propagating command: %w", err)
	}

	return nil
}

func (s *MasterServer) handleCommandPropagation(cmd command.Command) error {
	switch cmd.(type) {
	case command.Set:
		res, err := cmd.EncodedCommand()
		if err != nil {
			return fmt.Errorf("error encoding command: %w", err)
		}

		// Send the encoded command to all registered replica connections
		for _, replicaConn := range s.registeredReplicaConns {
			_, err := replicaConn.WriteString(res)
			if err != nil {
				return fmt.Errorf("error sending command to replica: %w", err)
			}
		}
	default:
		// this command does not need to be propagated
	}

	return nil
}

func (s *MasterServer) Run(ctx context.Context) error {
	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		func(clientConn connection.Connection, cmd command.Command) error {
			return s.ExecuteCommand(clientConn, cmd)
		},
	)
	go s.ConnectionHandler(ctx)
	go s.ExpiryLoop(ctx)

	return nil
}

func (s *MasterServer) CanHandleConnections() bool {
	return true
}

func (s *MasterServer) ShouldRespondToCommand(connection.Connection, command.Command) bool {
	return true
}
