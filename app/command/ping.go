package command

import (
	"fmt"
)

type Ping struct{}

func (Ping) String() string {
	return "PING"
}

func toPing(data []any) (Ping, error) {
	if len(data) != 0 {
		return Ping{}, fmt.Errorf("expected data for ping to be empty, but it was: %v", data)
	}
	return Ping{}, nil
}
