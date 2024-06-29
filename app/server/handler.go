package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"go.uber.org/zap"
)

const (
	// How often we should check in the background for expired keys
	expiryThreadPeriod = 10 * time.Second

	// How many keys we should check per expiry check
	samplesPerExpiry = 100

	// The size of the event queue. Note that a smaller number here can be used
	// in order to apply backpressure on the connectionHandlers
	eventQueueSize = 10
)

type Server struct {
	eventQueue chan Event
	listener   net.Listener

	// storeData is a map containing the keys and values held by this store
	storeData   map[string]storeValue
	storeDataMu *sync.Mutex

	Logger log.Logger
}

func NewServer(logger log.Logger, port int64) (Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return Server{}, fmt.Errorf("failed to bind to port %d: %w", port, err)
	}

	return Server{
		eventQueue:  make(chan Event, eventQueueSize),
		listener:    listener,
		Logger:      logger,
		storeData:   make(map[string]storeValue),
		storeDataMu: &sync.Mutex{},
	}, nil
}

func (s Server) EventLoop(ctx context.Context) {
	s.Logger.Info("starting event loop")
	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("event loop exiting", zap.Error(ctx.Err()))
			return
		case event := <-s.eventQueue:
			s.Logger.Info(
				"processing event",
				zap.String("command", event.Command),
				zap.Stringer("remoteAddress", event.ClientConn.RemoteAddr()),
			)

			parser, err := command.NewParser(event.Command, s.Logger)
			if err != nil {
				s.Logger.Error("error building parser from client command", zap.Error(err))
				continue
			}
			cmd, err := parser.Parse()
			if err != nil {
				s.Logger.Error("error parsing client command", zap.Error(err))
				continue
			}

			s.Logger.Info("executing command", zap.Stringer("command", cmd))

			responseStr, err := s.executeCommand(cmd)
			if err != nil {
				s.Logger.Error("error running client command", zap.Error(err))
				continue
			}

			s.Logger.Info(
				"sending response to client",
				zap.String("response", responseStr),
				zap.Stringer("remoteAddr", event.ClientConn.RemoteAddr()),
			)
			_, err = event.ClientConn.Write([]byte(responseStr))
			if err != nil {
				s.Logger.Error("error writing response to client", zap.Error(err), zap.String("response", responseStr))
				continue
			}
		}
	}
}

// ExpiryLoop will check a random sampling of at most `samplesPerExpiry` keys in the server's store to
// see if they are expired. Any found expired keys are deleted from the store
func (s Server) ExpiryLoop(ctx context.Context) {
	s.Logger.Info("starting expiry loop")
	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("event loop exiting", zap.Error(ctx.Err()))
			return
		case <-time.After(expiryThreadPeriod):
			s.storeDataMu.Lock()

			// NOTE: Map itterations in go are pseudo-random so
			// there's no need to explictly randomize this itteration
			inspectedKeys := int64(0)
			expiredKeys := int64(0)
			for key, value := range s.storeData {
				inspectedKeys++
				if value.isExpired() {
					expiredKeys++
					s.Logger.Debug(fmt.Sprintf("expiry loop deleting expired key %q", key))
					delete(s.storeData, key)
				}

				if inspectedKeys > samplesPerExpiry {
					break
				}
			}
			s.storeDataMu.Unlock()

			s.Logger.Info(
				"expiry loop completed a run",
				zap.Int64("inspectedKeys", inspectedKeys),
				zap.Int64("expiredKeys", expiredKeys),
			)
		}
	}
}

// ConnectionHandler listens for new pending connections and starts up a clientHandler goroutine for each new connection
func (s Server) ConnectionHandler(ctx context.Context) {
	s.Logger.Info(fmt.Sprintf("starting connection handler at %q", s.listener.Addr()))
	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("connection handler exiting", zap.Error(ctx.Err()))
			return
		default:
		}

		clientConn, err := s.listener.Accept()
		if err != nil {
			s.Logger.Error("error accepting connection", zap.Error(err))
			continue
		}

		go s.clientHandler(ctx, clientConn)
	}
}

// clienHandler is responsible for reading messages off of a connection and turning them into events
// which are then placedon the event queue
func (s Server) clientHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.Logger.Info(
		"starting client handler",
		zap.Stringer("remoteAddress", conn.RemoteAddr()),
	)

	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("client handler exiting", zap.Error(ctx.Err()))
			return
		default:
		}

		data := make([]byte, 512)
		bytesRead, err := conn.Read(data)
		if err != nil {
			s.Logger.Error("error reading from client connection", zap.Error(err))
			return
		}

		command := string(data[:bytesRead])
		s.Logger.Info("received command", zap.String("command", command))

		s.eventQueue <- Event{
			Command:    command,
			ClientConn: conn,
		}
	}
}
