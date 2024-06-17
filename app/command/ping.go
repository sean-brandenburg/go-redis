package command

import (
	"fmt"
	"net"
)

type Ping struct{}

func (_ Ping) String() string {
	return "Ping"
}

func (_ Ping) ExecuteCommand(conn net.Conn) error {
	_, err := conn.Write([]byte(fmt.Sprintf("+PONG%s", delimeter)))
	return err
}

func ToPing(data []string) (Ping, error) {
	if len(data) != 0 {
		return Ping{}, fmt.Errorf("expected data for ping to be empty, but it was: %v", data)
	}
	return Ping{}, nil
}
