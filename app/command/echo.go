package command

import (
	"fmt"
)

type Echo struct {
	Payload string
}

func (echo Echo) String() string {
	return fmt.Sprintf("ECHO: %q", echo.Payload)
}

func toEcho(data []any) (Echo, error) {
	if len(data) != 1 {
		return Echo{}, fmt.Errorf("expected only one data element for ping command, but found %d: %v", len(data), data)
	}
	res, ok := data[0].(string)
	if !ok {
		return Echo{}, fmt.Errorf("expected the input to the echo command to be a string but it was %[1]v of type %[1]v", data[0])
	}

	return Echo{res}, nil
}