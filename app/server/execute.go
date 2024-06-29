package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func (s *Server) executeCommand(cmd command.Command) (string, error) {
	switch typedCommand := cmd.(type) {
	case command.Ping:
		return s.executePing(typedCommand)
	case command.Echo:
		return s.executeEcho(typedCommand)
	case command.Get:
		return s.executeGet(typedCommand)
	case command.Set:
		return s.executeSet(typedCommand)
	}

	return "", fmt.Errorf("unknown command: %T", cmd)
}

func (s Server) executePing(_ command.Ping) (string, error) {
	res, err := command.Encode("PONG")
	if err != nil {
		return "", fmt.Errorf("error encoding response for PING command: %w", err)
	}
	return res, nil
}

func (s Server) executeEcho(echo command.Echo) (string, error) {
	res, err := command.Encode(echo.Payload)
	if err != nil {
		return "", fmt.Errorf("error encoding response for ECHO command: %w", err)
	}
	return res, nil
}

func (s Server) executeGet(get command.Get) (string, error) {
	data, ok := s.Get(get.Payload)
	if !ok {
		return command.NullBulkString, nil
	}

	res, err := command.Encode(data)
	if err != nil {
		return "", fmt.Errorf("error encoding response for GET command: %w", err)
	}
	return res, nil
}

func (s Server) executeSet(set command.Set) (string, error) {
	s.Set(set.KeyPayload, set.ValuePayload, set.ExpiryTime)
	res, err := command.Encode("OK")
	if err != nil {
		return "", fmt.Errorf("error encoding response for SET command: %w", err)
	}
	return res, nil
}
