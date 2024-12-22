package command

import (
	"fmt"
	"strings"
)

type Command interface {
	// String turns a command into a human readable string
	String() string

	// EncodedCommand turns a command into it's bulk string representation
	EncodedCommand() (string, error)

	CommandType() CommandType
}

type CommandType string

const (
	PingCmd     CommandType = "ping"
	EchoCmd     CommandType = "echo"
	InfoCmd     CommandType = "info"
	SetCmd      CommandType = "set"
	GetCmd      CommandType = "get"
	ReplConfCmd CommandType = "replconf"
	PSyncCmd    CommandType = "psync"
)

func ToCommand(data []any) (Command, error) {
	if len(data) == 0 {
		return nil, nil
	}

	rawCmd := data[0]
	cmdStr, ok := rawCmd.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert command %[1]v of type %[1]T to a string", rawCmd)
	}

	cmdType := strings.ToLower(cmdStr)
	cmdData := data[1:]

	switch CommandType(cmdType) {
	case PingCmd:
		return toPing(cmdData)
	case EchoCmd:
		return toEcho(cmdData)
	case InfoCmd:
		return toInfo(cmdData)
	case GetCmd:
		return toGet(cmdData)
	case SetCmd:
		return toSet(cmdData)
	case ReplConfCmd:
		return toReplConf(cmdData)
	case PSyncCmd:
		return toPSync(cmdData)
	default:
	}

	return nil, fmt.Errorf("unknown command %q", cmdType)
}
