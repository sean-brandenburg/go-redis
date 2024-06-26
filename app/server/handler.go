package server

import (
	"context"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"go.uber.org/zap"
)

type Server struct {
	Logger   log.Logger
	Events   chan Event
	Listener net.Listener

	// StoreData is a map containing the keys and values held by this store
	StoreData map[string]any
}

func (s Server) EventLoop(ctx context.Context) {
	s.Logger.Info("started event loop")
	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("event loop exiting", zap.Error(ctx.Err()))
			return
		case event := <-s.Events:
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

			s.Logger.Info("sending response to client", zap.String("response", responseStr))
			_, err = event.ClientConn.Write([]byte(responseStr))
			if err != nil {
				s.Logger.Error("error writing response to client", zap.Error(err), zap.String("response", responseStr))
				continue
			}
		}
	}
}

func (s Server) ConnectionHandler(ctx context.Context) {
	s.Logger.Info("started connection handler")
	for {
		select {
		case <-ctx.Done():
			s.Logger.Error("connection handler exiting", zap.Error(ctx.Err()))
			return
		default:
		}

		clientConn, err := s.Listener.Accept()
		if err != nil {
			s.Logger.Error("error accepting connection", zap.Error(err))
			continue
		}

		go s.clientHandler(ctx, clientConn)
	}
}

func (s Server) clientHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()

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

		s.Events <- Event{
			Command:    command,
			ClientConn: conn,
		}
	}
}
