package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"go.uber.org/zap"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/connection"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type ReplicaServer struct {
	BaseServer

	// The `hostname:port` string that should be used to connect
	// to the master of this replica's replica set
	masterAddress string

	masterConnection connection.Connection

	steadyState bool

	shouldIgnoreMaster bool
}

func NewReplicaServer(logger log.Logger, masterAddress string, opts ServerOptions) (ReplicaServer, error) {
	baseServer, err := NewBaseServer(logger, opts)
	if err != nil {
		return ReplicaServer{}, fmt.Errorf("error initializing replica server: %w", err)
	}

	return ReplicaServer{
		BaseServer:         baseServer,
		masterAddress:      masterAddress,
		steadyState:        false,
		shouldIgnoreMaster: false,
	}, nil
}

func (s *ReplicaServer) NodeType() NodeType {
	return ReplicaNodeType
}

func (s *ReplicaServer) Run(ctx context.Context) error {
	conn, err := net.Dial("tcp", s.masterAddress)
	if err != nil {
		return fmt.Errorf("failed to dial master at address %q: %s", s.masterAddress, err)
	}
	s.masterConnection = connection.NewNetworkConn(conn, connection.MasterConnection, s.logger)

	// 0. Start the connection handler for the replica before we sync with the master
	// so that we can accept connections. We will not read requests from these connections until we're in steady state
	go s.ExpiryLoop(ctx)
	go s.ConnectionHandler(ctx)
	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		func(conn connection.Connection, cmd command.Command) error {
			return s.ExecuteCommand(conn, cmd)
		},
	)

	// 1. The replica sends a ping to it's master
	res, err := s.SendCommandToMaster(ctx, &command.Ping{})
	if err != nil {
		return fmt.Errorf("failed to PING master at address %q: %s", s.masterAddress, err)
	}
	if res != "+PONG\r\n" {
		return fmt.Errorf("unexpected response to PING to master at address %q: %s", s.masterAddress, err)
	}

	// 2. The replica sends it's port as a REPLCONF
	res, err = s.SendCommandToMaster(ctx, &command.ReplConf{Payload: []string{"listening-port", strconv.Itoa(s.listenerPort)}})
	if err != nil {
		return fmt.Errorf("failed to send first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}
	if res != command.OKString {
		return fmt.Errorf("unexpected response to first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}

	// 3. The replica sends it's capabilities
	res, err = s.SendCommandToMaster(ctx, &command.ReplConf{Payload: []string{"capa", "psync2"}})
	if err != nil {
		return fmt.Errorf("failed to send first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}
	if res != command.OKString {
		return fmt.Errorf("unexpected response to first REPLCONF to master at address %q: %s", s.masterAddress, err)
	}

	// 4a. The replica sends a PSYNC to master to get a replicationID
	res, err = s.SendCommandToMaster(ctx, &command.PSync{ReplicationID: "?", MasterOffset: "-1"})
	if err != nil {
		return fmt.Errorf("failed to send PSYNC message to master: %w", err)
	}
	s.Logger().Info("received response to PSYNC command", zap.String("response", res))

	// 4b. Read off the RDB file
	res, err = s.masterConnection.ReadRDBFile()
	if err != nil {
		return fmt.Errorf("failed to read off RDB file after sending PSYNC message: %w", err)
	}
	s.Logger().Info("received RDB data", zap.String("data", res))

	// 5. Start up client handler for the master conn and set the replica to steady state
	go s.clientHandler(ctx, s.masterConnection)
	s.SetIsSteadyState(true)

	return nil
}

func (s *ReplicaServer) ExecuteCommand(conn connection.Connection, cmd command.Command) error {
	s.Logger().Info(fmt.Sprintf("replica executing command: %v", cmd))

	err := RunCommand(s, conn, cmd)
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

	_, err = s.masterConnection.WriteString(encodedCmd)
	if err != nil {
		return "", fmt.Errorf("failed to write command %q to master at address %q: %s", encodedCmd, s.masterAddress, err)
	}

	// TODO: Add generic reaad command from conn function that can be used here and elsewhere
	responseCmdStr, err := s.masterConnection.ReadNextCmdString()
	if err != nil {
		return "", fmt.Errorf("failed to read response for command from master at address %q: %s", s.masterAddress, err)
	}

	s.logger.Info(
		"successfully sent command to master node",
		zap.String("masterAddress", s.masterAddress),
		zap.String("command", encodedCmd),
	)

	return responseCmdStr, nil
}

func (s *ReplicaServer) CanHandleConnections() bool {
	return s.steadyState
}

func (s *ReplicaServer) SetIsSteadyState(steadyState bool) {
	s.steadyState = steadyState
}

func (s *ReplicaServer) ShouldRespondToCommand(conn connection.Connection, cmd command.Command) bool {
	return !s.shouldIgnoreMaster ||
		conn.ConnectionType() != connection.MasterConnection ||
		cmd.CommandType() == command.ReplConfCmd

}

func (s *ReplicaServer) SetShouldIgnoreMaster(shouldIgnoreMaster bool) {
	s.shouldIgnoreMaster = shouldIgnoreMaster
}
