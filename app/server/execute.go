package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func executeCommand(server Server, cmd command.Command) (string, error) {
	switch typedCommand := cmd.(type) {
	case command.Ping:
		return executePing(typedCommand)
	case command.Echo:
		return executeEcho(typedCommand)
	case command.Info:
		return executeInfo(server, typedCommand)
	case command.Get:
		return executeGet(server, typedCommand)
	case command.Set:
		return executeSet(server, typedCommand)
	}

	return "", fmt.Errorf("unknown command: %T", cmd)
}

func executePing(_ command.Ping) (string, error) {
	return "+PONG\r\n", nil
}

func executeEcho(echo command.Echo) (string, error) {
	res, err := command.Encoder{}.Encode(echo.Payload)
	if err != nil {
		return "", fmt.Errorf("error encoding response for ECHO command: %w", err)
	}
	return res, nil
}

func executeGet(server Server, get command.Get) (string, error) {
	data, ok := server.Get(get.Payload)
	if !ok {
		return command.NullBulkString, nil
	}

	res, err := command.Encoder{}.Encode(data)
	if err != nil {
		return "", fmt.Errorf("error encoding response for GET command: %w", err)
	}
	return res, nil
}

func executeSet(server Server, set command.Set) (string, error) {
	server.Set(set.KeyPayload, set.ValuePayload, set.ExpiryTimeMs)
	return command.OKString, nil
}

// TODO: Testing once fn returns are a bit more stable
func executeInfo(server Server, info command.Info) (string, error) {
	serverInfo, err := GetServerInfo(server, info.Payload)
	if err != nil {
		return "", fmt.Errorf("error executing info command %q: %w", info, err)
	}

	// Sort and join info with new lines
	infoToEncode := make([]string, 0, len(serverInfo))
	for key, val := range serverInfo {
		infoToEncode = append(infoToEncode, fmt.Sprintf("%s:%s\n", key, val))
	}
	slices.Sort(infoToEncode)

	encoder := command.Encoder{UseBulkStrings: true}
	res, err := encoder.Encode(strings.Join(infoToEncode, ""))
	if err != nil {
		return "", fmt.Errorf("error encoding response for INFO command: %w", err)
	}
	return res, nil
}
