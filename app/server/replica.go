package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type ReplicaServer struct {
	BaseServer

	// The `hostname:port` string that should be used to connect
	// to the master of this replica's replica set
	masterAddress string

	masterConnection net.Conn

	steadyState bool
}

func NewReplicaServer(logger log.Logger, masterAddress string, opts ServerOptions) (ReplicaServer, error) {
	baseServer, err := NewBaseServer(logger, opts)
	if err != nil {
		return ReplicaServer{}, fmt.Errorf("error initializing replica server: %w", err)
	}

	return ReplicaServer{
		BaseServer:    baseServer,
		masterAddress: masterAddress,
		steadyState:   false,
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
	s.masterConnection = conn

	// 0. Start the event/connection handling for the replica before we sync with the master
	// since the tests expect that we are immediately ready to start handling connections after we respond
	//
	// In the real world though, I would want this to be after we've synchronized with the mater so that a client won't
	// be able to connect and read before we have a consistent state
	go EventLoop(
		ctx,
		s.logger,
		s.eventQueue,
		func(conn Connection, cmd command.Command) error {
			return s.ExecuteCommand(conn, cmd)
		},
	)
	go s.ConnectionHandler(ctx)

	go s.ExpiryLoop(ctx)

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
	psync := command.PSync{ReplicationID: "?", MasterOffset: "-1"}
	encodedCmd, err := psync.EncodedCommand()
	if err != nil {
		return fmt.Errorf("failed to encode command %v: %s", psync, err)
	}
	_, err = s.masterConnection.Write([]byte(encodedCmd))
	if err != nil {
		return fmt.Errorf("failed to write command %q to master at address %q: %s", encodedCmd, s.masterAddress, err)
	}

	// 4b. Read off the FULLRESYNC
	r := bufio.NewReader(s.masterConnection)
	fullResyncMsg, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read PSYNC response FULLRESYNC for command from master at address %q: %s", s.masterAddress, err)
	}
	s.Logger().Info("received response to PSYNC command", zap.String("response", fullResyncMsg))

	// 4c. Read off the size of the RDB file. Should get a response like $88\r\n
	RDBFileSize, err := r.ReadString('\n')
	s.Logger().Info("received RDB file size string", zap.String("data", RDBFileSize))

	if err != nil {
		return fmt.Errorf("failed to read PSYNC response FULLRESYNC for command from master at address %q: %s", s.masterAddress, err)
	}

	// 4d. String should now look like $88
	trimmedRDBFileSizeStr := strings.TrimSuffix(RDBFileSize, "\r\n")
	bytesToRead, err := command.ParseIntWithPrefix(trimmedRDBFileSizeStr, "$")
	if err != nil {
		return fmt.Errorf("RDB file response to PSYNC did not contain an int size in its header: %q", trimmedRDBFileSizeStr)
	}

	// 4e. Read the expected number of bytes
	buf := make([]byte, bytesToRead)
	rdbData, err := r.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading RDB data: %w", err)
	}
	s.Logger().Info("received RDB data", zap.String("data", string(buf[:rdbData])))

	// 5. Start up client handler for the master conn (TODO: The master expects this to already be up at this point sometimes)
	go s.clientHandler(ctx, ConnWithType{Conn: s.masterConnection, ConnType: MasterConnection})

	return nil
}

func (s *ReplicaServer) ExecuteCommand(conn Connection, cmd command.Command) error {
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

	_, err = s.masterConnection.Write([]byte(encodedCmd))
	if err != nil {
		return "", fmt.Errorf("failed to write command %q to master at address %q: %s", encodedCmd, s.masterAddress, err)
	}

	rawRes := make([]byte, MaxMessageSize)
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

func (s *ReplicaServer) IsSteadyState() bool {
	return s.steadyState
}

func (s *ReplicaServer) SetIsSteadyState(steadyState bool) {
	s.steadyState = steadyState
}
