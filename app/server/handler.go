package server

import (
	"context"
	"fmt"
	"net"
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

	// Max bytes in a message
	MaxMessageSize = 512
)

type ExecuteCommand func(clientConn net.Conn, cmd command.Command) error

func EventLoop(ctx context.Context, logger log.Logger, eventQueue chan Event, server Server) {
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
				zap.Stringer("remoteAddress", event.ClientConn.RemoteAddr()),
			)

			parser, err := command.NewParser(event.Command, logger)
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
			err = server.ExecuteCommand(event.ClientConn, cmd)
			if err != nil {
				logger.Error("error executing client command", zap.Error(err))
				continue
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
	s.logger.Info(fmt.Sprintf("starting connection handler at %q", s.listener.Addr()))
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

		go s.clientHandler(ctx, clientConn)
	}
}

// clienHandler is responsible for reading messages off of a connection and turning them into events
// which are then placedon the event queue
func (s BaseServer) clientHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.logger.Info(
		"starting client handler",
		zap.Stringer("remoteAddress", conn.RemoteAddr()),
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Error("client handler exiting", zap.Error(ctx.Err()))
			return
		default:
		}

		data := make([]byte, MaxMessageSize)
		bytesRead, err := conn.Read(data)
		if err != nil {
			s.logger.Error("error reading from client connection", zap.Error(err))
			return
		}

		command := string(data[:bytesRead])
		s.logger.Info("received command", zap.String("command", command))

		s.eventQueue <- Event{
			Command:    command,
			ClientConn: conn,
		}
	}
}

func GetServerInfo(server Server, infoType string) (map[string]string, error) {
	if infoType != "replication" {
		return nil, fmt.Errorf("received unexpected info type %q", infoType)
	}

	return map[string]string{
		"role":               server.NodeType(),
		"master_repl_offset": "0",
		"master_replid":      command.HARDCODEC_REPL_ID,
	}, nil
}
