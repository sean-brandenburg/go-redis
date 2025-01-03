package command

import (
	"fmt"
)

type Ping struct{}

func (Ping) String() string {
	return "PING"
}

func (Ping) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	return e.EncodeArray([]any{string(PingCmd)})
}

func (Ping) CommandType() CommandType {
	return PingCmd
}

func toPing(data []any) (Ping, error) {
	if len(data) != 0 {
		return Ping{}, fmt.Errorf("expected data for PING to be empty, but it was: %v", data)
	}
	return Ping{}, nil
}
