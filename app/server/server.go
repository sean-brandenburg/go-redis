package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type Server interface {
	// Run the server
	Run(ctx context.Context) error

	// Set sets a key in the server's store
	Set(key string, value any, expiryTimeMs int64)

	// Get's a value from the server's store and returns a bool
	// indicating whether or not the key was found
	Get(key string) (any, bool)

	NodeType() string
}

type BaseServer struct {
	eventQueue chan Event
	listener   net.Listener

	// storeData is a map containing the keys and values held by this store
	storeData   map[string]storeValue
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
		eventQueue:  make(chan Event, eventQueueSize),
		listener:    listener,
		logger:      logger,
		storeData:   make(map[string]storeValue),
		storeDataMu: &sync.Mutex{},
	}, nil
}

func (s *BaseServer) Logger() log.Logger {
	return s.logger
}

func (s *BaseServer) NodeType() string {
	return "base"
}

func (s *BaseServer) Run(ctx context.Context) error {
	return fmt.Errorf("the base server's run should not be used and exists only to fulfill the Server interface to simplify testing")
}
