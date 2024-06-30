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
}

type CommandType string

const (
	pingCmd     CommandType = "ping"
	echoCmd     CommandType = "echo"
	infoCmd     CommandType = "info"
	setCmd      CommandType = "set"
	getCmd      CommandType = "get"
	replConfCmd CommandType = "replconf"
	pSyncCmd    CommandType = "psync"
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
	case pingCmd:
		return toPing(cmdData)
	case echoCmd:
		return toEcho(cmdData)
	case infoCmd:
		return toInfo(cmdData)
	case getCmd:
		return toGet(cmdData)
	case setCmd:
		return toSet(cmdData)
	case replConfCmd:
		return toReplConf(cmdData)
	case pSyncCmd:
		return toPSync(cmdData)
	default:
	}

	return nil, fmt.Errorf("unknown command %q", cmdType)
}
