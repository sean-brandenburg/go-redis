package server

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/connection"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

const (
	// How often we should check in the background for expired keys
	expiryThreadPeriod = 10 * time.Second

	// How many keys we should check per expiry check
	samplesPerExpiry = 100

	// The size of the event queue. Note that a smaller number here can be used
	// in order to apply backpressure on the connectionHandlers
	eventQueueSize = 10

	// Max bytes in a message
	MaxMessageSize = 1024
)

type ExecuteCommand func(conn connection.Connection, cmd command.Command) error

type Event struct {
	// The event string to be handled
	Command string

	// The client connection that this event came from
	Conn connection.Connection
}

func EventLoop(ctx context.Context, logger log.Logger, eventQueue chan Event, execute ExecuteCommand) {
	logger.Info("starting event loop")
	for {
		select {
		case <-ctx.Done():
			logger.Error("event loop exiting", zap.Error(ctx.Err()))
			return
		case event := <-eventQueue:
			logger.Info(
				"processing event",
				zap.String("command", event.Command),
				zap.Stringer("remoteAddress", event.Conn.RemoteAddr()),
			)

			parser, err := command.NewParser(event.Command)
			if err != nil {
				logger.Error("error building parser from client command", zap.Error(err))
				continue
			}
			cmd, err := parser.Parse()
			if err != nil {
				logger.Error("error parsing client command", zap.Error(err))
				continue
			}

			logger.Info("executing command", zap.Stringer("command", cmd))

			err = execute(event.Conn, cmd)
			if err != nil {
				logger.Error("error executing client command, skipping execution", zap.Error(err))
			}
		}
	}
}

// ExpiryLoop will check a random sampling of at most `samplesPerExpiry` keys in the server's store to
// see if they are expired. Any found expired keys are deleted from the store
func (s BaseServer) ExpiryLoop(ctx context.Context) {
	s.logger.Info("starting expiry loop")
	for {
		select {
		case <-ctx.Done():
			s.logger.Error("event loop exiting", zap.Error(ctx.Err()))
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
					s.logger.Debug(fmt.Sprintf("expiry loop deleting expired key %q", key))
					delete(s.storeData, key)
				}

				if inspectedKeys > samplesPerExpiry {
					break
				}
			}
			s.storeDataMu.Unlock()

			s.logger.Info(
				"expiry loop completed a run",
				zap.Int64("inspectedKeys", inspectedKeys),
				zap.Int64("expiredKeys", expiredKeys),
			)
		}
	}
}

// ConnectionHandler listens for new pending connections and starts up a clientHandler goroutine for each new connection
func (s BaseServer) ConnectionHandler(ctx context.Context) {
	s.logger.Info("starting connection handler at %q", zap.Stringer("connectionAddr", s.listener.Addr()))
	for {
		select {
		case <-ctx.Done():
			s.logger.Error("connection handler exiting", zap.Error(ctx.Err()))
			return
		default:
		}

		clientConn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("error accepting connection", zap.Error(err))
			continue
		}

		s.logger.Info("accepted connection from client", zap.Stringer("remoteAddress", clientConn.RemoteAddr()))

		go s.clientHandler(ctx, connection.NewNetworkConn(clientConn, connection.ClientConnection, s.logger))
	}
}

func (s BaseServer) waitUntilCanHandleConnections(ctx context.Context) error {
	if s.CanHandleConnections() {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond * 200):
			if s.CanHandleConnections() {
				return nil
			}
		}
	}
}

// clienHandler is responsible for reading messages off of a connection and turning them into events
// which are then placed on the event queue
func (s BaseServer) clientHandler(ctx context.Context, conn connection.Connection) {
	defer conn.Close()

	err := s.waitUntilCanHandleConnections(ctx)
	if err != nil {
		s.logger.Error("failed to wait until client can be handled", zap.Error(err))
		return
	}

	s.logger.Info("starting client handler", zap.Stringer("remoteAddress", conn.RemoteAddr()))

	for {
		select {
		case <-ctx.Done():
			s.logger.Error("client handler exiting", zap.Error(ctx.Err()))
			return
		default:
			command, err := conn.ReadNextCmdString()
			if err != nil {
				s.logger.Error("error reading next command from client connection", zap.Error(err))
				return
			}
			s.logger.Info("received command", zap.String("command", command))

			s.eventQueue <- Event{
				Command: command,
				Conn:    conn,
			}
		}
	}
}

func GetServerInfo(server Server, infoType string) (map[string]string, error) {
	if infoType != "replication" {
		return nil, fmt.Errorf("received unexpected info type %q", infoType)
	}

	return map[string]string{
		"role":               string(server.NodeType()),
		"master_repl_offset": "0",
		"master_replid":      command.HARDCODE_REPL_ID,
	}, nil
}
