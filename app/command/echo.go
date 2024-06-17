package command

import (
	"fmt"
	"net"
)

type Echo struct {
	Payload string
}

func (cmd Echo) String() string {
	return fmt.Sprintf("Echo: %q", cmd.Payload)
}

func (cmd Echo) ExecuteCommand(conn net.Conn) error {
	fmt.Println("payload: ", cmd.Payload)
	_, err := conn.Write([]byte(fmt.Sprintf("+%s%s", cmd.Payload, delimeter)))
	return err
}

func ToEcho(data []string) (Echo, error) {
	if len(data) != 2 {
		return Echo{}, fmt.Errorf("Expected 2 data entries to follow echo command but got %d entries: %v", len(data), data)
	}

	dataLen, err := parseIntWithPrefix(data[0], "$")
	if err != nil {
		return Echo{}, err
	}

	echoData := data[1]
	if dataLen != len(echoData) {
		return Echo{}, fmt.Errorf("Expected data to be of length %d, but it had length %d: %q", dataLen, len(echoData), echoData)
	}

	return Echo{echoData}, nil
}
