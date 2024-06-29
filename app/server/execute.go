package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func (s *Server) executeCommand(cmd command.Command) (string, error) {
	switch typedCommand := cmd.(type) {
	case command.Ping:
		return s.executePing(typedCommand)
	case command.Echo:
		return s.executeEcho(typedCommand)
	case command.Info:
		return s.executeInfo(typedCommand)
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
	s.Set(set.KeyPayload, set.ValuePayload, set.ExpiryTimeMs)
	res, err := command.Encode("OK")
	if err != nil {
		return "", fmt.Errorf("error encoding response for SET command: %w", err)
	}
	return res, nil
}

func (s Server) executeInfo(info command.Info) (string, error) {
	serverInfo, err := s.getInfo(info.Payload)
	if err != nil {
		return "", fmt.Errorf("error executing info command %q: %w", info, err)
	}

	// Sort and join info with new lines
	infoToEncode := make([]string, 0, len(serverInfo))
	for key, val := range serverInfo {
		infoToEncode = append(infoToEncode, fmt.Sprintf("%s:%s", key, val))
	}
	slices.Sort(infoToEncode)

	res, err := command.Encode(strings.Join(infoToEncode, "\n"))
	if err != nil {
		return "", fmt.Errorf("error encoding response for INFO command: %w", err)
	}
	return res, nil
}
