package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type NodeType string

const (
	MasterNodeType  NodeType = "master"
	ReplicaNodeType NodeType = "slave"
	BaseNodeType    NodeType = "base"
)

type Server interface {
	// Run the server
	Run(ctx context.Context) error

	// ExecuteCommand runs a command on this server
	ExecuteCommand(conn Connection, command command.Command) error

	// Set sets a key in the server's store
	Set(key string, value any, expiryTimeMs int64)

	// Get fetches a value from the server's store and returns a bool
	// indicating whether or not the key was found
	Get(key string) (any, bool)

	// Size Returns the number of items in the store
	Size() int

	// Logger returns this server's logger
	Logger() log.Logger

	// NodeType returns the type of this server
	NodeType() NodeType

	// CanHandleConnections is true if a server is up and ready to handle events from connections
	CanHandleConnections() bool

	// ShouldRespondToCommand is true if this server should send a response after processing a message from a Connection
	ShouldRespondToCommand(Connection, command.Command) bool
}

type BaseServer struct {
	eventQueue   chan Event
	listener     net.Listener
	listenerPort int

	// storeData is a map containing the keys and values held by this store
	storeData   serverStore
	storeDataMu *sync.Mutex

	logger log.Logger
}

type ServerOptions struct {
	Port *int
}

func NewBaseServer(logger log.Logger, opts ServerOptions) (BaseServer, error) {
	port := 6379

	if opts.Port != nil {
		port = *opts.Port
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return BaseServer{}, fmt.Errorf("failed to bind to port %d: %w", port, err)
	}

	return BaseServer{
		eventQueue:   make(chan Event, eventQueueSize),
		listener:     listener,
		listenerPort: port,
		logger:       logger,
		storeData:    make(map[string]storeValue),
		storeDataMu:  &sync.Mutex{},
	}, nil
}

func (s *BaseServer) Logger() log.Logger {
	return s.logger
}

func (s *BaseServer) NodeType() NodeType {
	return BaseNodeType
}

// NOTE: The base server implementation of ExecuteCommand should only be used in tests
// Otherwise we should use the MasterServer and ReplicaServer implementations
func (s *BaseServer) ExecuteCommand(conn Connection, command command.Command) error {
	return RunCommand(s, conn, command)
}

func (s *BaseServer) Run(ctx context.Context) error {
	return fmt.Errorf("the base server's run should not be used and exists only to fulfill the Server interface to simplify testing")
}

func (s *BaseServer) CanHandleConnections() bool {
	return true
}

func (s *BaseServer) ShouldRespondToCommand(Connection, command.Command) bool {
	return true
}
