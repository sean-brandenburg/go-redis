package command

import (
	"fmt"
	"strings"
)

type Command interface {
	String() string
}

type CommandType string

const (
	pingCmd CommandType = "ping"
	echoCmd CommandType = "echo"
	setCmd  CommandType = "set"
	getCmd  CommandType = "get"
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
		return ToPing(cmdData)
	case echoCmd:
		return ToEcho(cmdData)
	case getCmd:
		return ToGet(cmdData)
	case setCmd:
		return ToSet(cmdData)
	default:
	}

	return nil, fmt.Errorf("unknown command %q", cmdType)
}
