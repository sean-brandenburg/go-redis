package command

import (
	"fmt"
	"strconv"
	"strings"
)

type ReplConf struct {
	KeyPayload   string
	ValuePayload string
}

func (conf ReplConf) String() string {
	return fmt.Sprintf("REPLCONF: %q -> %q", conf.KeyPayload, conf.ValuePayload)
}

func (conf ReplConf) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	return e.Encode([]any{string(replConfCmd), conf.KeyPayload, conf.ValuePayload})
}

func (conf ReplConf) IsGetAck() bool {
	return strings.ToLower(conf.KeyPayload) == "getack" && conf.ValuePayload == "*"
}

func (conf ReplConf) IsListeningPort() bool {
	return strings.ToLower(conf.KeyPayload) == "listening-port"
}

func (conf ReplConf) IsAck() bool {
	if _, err := strconv.ParseInt(conf.ValuePayload, 10, 64); err == nil {
		return strings.ToLower(conf.KeyPayload) == "ack"
	}
	return false
}

func toReplConf(data []any) (ReplConf, error) {
	if len(data) != 2 {
		return ReplConf{}, fmt.Errorf("expected exactly 2 data elements for REPLCONF command, but found %d: %v", len(data), data)
	}

	key, ok := data[0].(string)
	if !ok {
		return ReplConf{}, fmt.Errorf("expected the key input to the REPLCONF command to be a string but it was %[1]v of type %[1]v", data[0])
	}

	value, ok := data[1].(string)
	if !ok {
		return ReplConf{}, fmt.Errorf("expected the value input to the REPLCONF command to be a string but it was %[1]v of type %[1]v", data[1])
	}

	return ReplConf{
		KeyPayload:   key,
		ValuePayload: value,
	}, nil
}
