package server

import (
	"context"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"go.uber.org/zap"
)

type ReplicaServer struct {
	BaseServer

	// The `hostname:port` string that should be used to connect
	// to the master of this replica's replica set
	masterAddress string

	masterConnection net.Conn
}

func NewReplicaServer(logger log.Logger, masterAddress string, opts ServerOptions) (ReplicaServer, error) {
	baseServer, err := NewBaseServer(logger, opts)
	if err != nil {
		return ReplicaServer{}, fmt.Errorf("error initializing replica server: %w", err)
	}

	return ReplicaServer{
		BaseServer:    baseServer,
		masterAddress: masterAddress,
	}, nil
}

func (s *ReplicaServer) NodeType() string {
	return "slave"
}

func (s *ReplicaServer) Run(ctx context.Context) error {
	conn, err := net.Dial("tcp", s.masterAddress)
	if err != nil {
		return fmt.Errorf("failed to dial master at address %q: %s", s.masterAddress, err)
	}
	s.masterConnection = conn

	// 0. The replica sends a ping to it's master
	res, err := s.SendCommandToMaster(ctx, &command.Ping{})
	if err != nil {
		return fmt.Errorf("failed to PING master at address %q: %s", s.masterAddress, err)
	}
	if res != "+PONG\r\n" {
		return fmt.Errorf("unexpected response to PING to master at address %q: %s", s.masterAddress, err)
	}

	// 1. The replica sends it's port as a REPLCONF
	res, err = s.SendCommandToMaster(ctx, &command.ReplConf{KeyPayload: "listening-port", ValuePayload: fmt.Sprintf("%d", s.listenerPort)})
	if err != nil {
		return fmt.Errorf("failed to send first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}
	if res != command.OKString {
		return fmt.Errorf("unexpected response to first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}

	// 2. The replica sends it's capabilities
	res, err = s.SendCommandToMaster(ctx, &command.ReplConf{KeyPayload: "capa", ValuePayload: "psync2"})
	if err != nil {
		return fmt.Errorf("failed to send first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}
	if res != command.OKString {
		return fmt.Errorf("unexpected response to first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}

	// 3. The replica sends a PSYNC to master to get a replicationID
	strRes, err := s.SendCommandToMaster(ctx, &command.PSync{ReplicationID: "?", MasterOffset: "-1"})
	if err != nil {
		return fmt.Errorf("failed to send PSYNC to master at address %q: %s", s.masterAddress, err)
	}

	s.Logger().Info("received response to PSYNC command", zap.String("response", strRes))
	// TODO: Parse the response to this cmd
	// if res != "" {
	// 	return fmt.Errorf("unexpected response to PSYNC to master at address %q: %s", s.masterAddress, err)
	// }

	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		func(clientConn net.Conn, cmd command.Command) error {
			return s.ExecuteCommand(clientConn, cmd)
		},
	)
	go s.ConnectionHandler(ctx)

	// Don't send responses back to the master since it's just propagating messages
	go s.clientHandler(ctx, s.masterConnection, false)

	go s.ExpiryLoop(ctx)

	return nil
}

func (s *ReplicaServer) ExecuteCommand(clientConn net.Conn, cmd command.Command) error {
	s.Logger().Info(fmt.Sprintf("replica executing command: %v", cmd))
	err := commandExecutor{
		server:     s,
		clientConn: clientConn,
	}.execute(cmd)
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	return nil
}

func (s *ReplicaServer) SendCommandToMaster(ctx context.Context, cmd command.Command) (string, error) {
	encodedCmd, err := cmd.EncodedCommand()
	if err != nil {
		return "", fmt.Errorf("failed to encode command %v: %s", cmd, err)
	}

	_, err = s.masterConnection.Write([]byte(encodedCmd))
	if err != nil {
		return "", fmt.Errorf("failed to write command %q to master at address %q: %s", encodedCmd, s.masterAddress, err)
	}

	rawRes := make([]byte, 512)
	bytesRead, err := s.masterConnection.Read(rawRes)
	if err != nil {
		return "", fmt.Errorf("failed to read response for command from master at address %q: %s", s.masterAddress, err)
	}
	strRes := string(rawRes[:bytesRead])

	s.logger.Info(
		"successfully sent command to master node",
		zap.String("masterAddress", s.masterAddress),
		zap.String("command", encodedCmd),
	)

	return strRes, nil
}
