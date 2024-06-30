package command

import (
	"fmt"
)

type Info struct {
	Payload string
}

func (info Info) String() string {
	return fmt.Sprintf("INFO: %q", info.Payload)
}

func (info Info) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	return e.Encode([]any{string(infoCmd), info.Payload})
}

func toInfo(data []any) (Info, error) {
	if len(data) != 1 {
		return Info{}, fmt.Errorf("expected only one data element for INFO command, but found %d: %v", len(data), data)
	}
	res, ok := data[0].(string)
	if !ok {
		return Info{}, fmt.Errorf("expected the input to the echo command to be a string but it was %[1]v of type %[1]v", data[0])
	}

	// At this point, 'replication' is the only valid value
	if res != "replication" {
		return Info{}, fmt.Errorf("expected the input for the INFO command to be 'replication' but it was %q", res)
	}

	return Info{Payload: res}, nil
}
