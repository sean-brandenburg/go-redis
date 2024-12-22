package command

import (
	"fmt"
	"strconv"
	"strings"
)

type ReplConf struct {
	Payload []string
}

func (conf ReplConf) String() string {
	return fmt.Sprintf("REPLCONF: %v", conf.Payload)
}

func (conf ReplConf) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	toEncode := []any{string(ReplConfCmd)}
	for _, payload := range conf.Payload {
		toEncode = append(toEncode, payload)
	}
	return e.EncodeArray(toEncode)
}

func (ReplConf) CommandType() CommandType {
	return ReplConfCmd
}

func (conf ReplConf) IsGetAck() bool {
	if len(conf.Payload) != 2 {
		return false
	}

	return strings.ToLower(conf.Payload[0]) == "getack" && conf.Payload[1] == "*"
}

func (conf ReplConf) IsListeningPort() bool {
	if len(conf.Payload) != 2 {
		return false
	}

	return strings.ToLower(conf.Payload[0]) == "listening-port"
}

func (conf ReplConf) IsAck() bool {
	if len(conf.Payload) != 2 {
		return false
	}

	if _, err := strconv.ParseInt(conf.Payload[1], 10, 64); err == nil {
		return strings.ToLower(conf.Payload[0]) == "ack"
	}
	return false
}

func (conf ReplConf) IsCapa() bool {
	if len(conf.Payload) < 2 && len(conf.Payload)%2 != 0 {
		return false
	}

	for idx, val := range conf.Payload {
		if idx%2 == 0 && strings.ToLower(val) != "capa" {
			return false
		}
	}

	return true
}

func toReplConf(data []any) (ReplConf, error) {
	res := ReplConf{Payload: []string{}}
	for _, elem := range data {
		value, ok := elem.(string)
		if !ok {
			return ReplConf{}, fmt.Errorf("expected the key input to the REPLCONF command to be a string but it was %[1]v of type %[1]v", data[0])
		}
		res.Payload = append(res.Payload, value)
	}

	return res, nil
}
