package command

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Command interface {
	String() string
	ExecuteCommand(net.Conn) error
}

type CommandType string

const (
	pingCmd CommandType = "ping"
	echoCmd CommandType = "echo"
)

const delimeter = "\r\n"

func ParseCommand(cmd string) (Command, error) {
	cmdStrs := strings.Split(
		strings.TrimSuffix(cmd, delimeter),
		delimeter,
	)
	if len(cmdStrs) == 0 {
		return nil, nil
	}

	// 0. Parse out the number of parameters that we expect
	numParams, err := parseIntWithPrefix(cmdStrs[0], "*")
	if err != nil {
		return nil, fmt.Errorf("Expected first param to be the number of params but got: %q", cmdStrs[0])
	}
	if numParams == 0 {
		// Valid empty command: $0\r\n\r\n
		if len(cmdStrs) == 2 && cmdStrs[1] == "" {
			return nil, nil
		}
		// Otherwise this command is not valid
		return nil, fmt.Errorf("Malformed command string: %q", cmd)
	}

	// 1. Parse out the command that we're running and it's params
	if len(cmdStrs) < 3 {
		return nil, fmt.Errorf("Not enough parameters provided in command: %q", cmd)
	}
	err = validateLengthDataPair(cmdStrs[1], cmdStrs[2])
	if err != nil {
		return nil, fmt.Errorf("Error validating command type: %w", err)
	}
	cmdType := strings.ToLower(cmdStrs[2])
	cmdParams := cmdStrs[3:]

	// 2. Return the relevant command
	switch CommandType(cmdType) {
	case pingCmd:
		return ToPing(cmdParams)
	case echoCmd:
		return ToEcho(cmdParams)
	default:
	}

	return nil, fmt.Errorf("unknown command type: ", cmdType)
}

func parseIntWithPrefix(intStr string, prefix string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(intStr, prefix))
}

func validateLengthDataPair(lengthStr string, dataStr string) error {
	length, err := parseIntWithPrefix(lengthStr, "$")
	if err != nil {
		return err
	}

	if length != len(dataStr) {
		return fmt.Errorf("length of %d does not match the length of the provided data %q", length, dataStr)
	}
	return nil
}
