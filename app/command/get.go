package command

import (
	"fmt"
)

type Get struct {
	Payload string
}

func (get Get) String() string {
	return fmt.Sprintf("GET: %q", get.Payload)
}

func toGet(data []any) (Get, error) {
	if len(data) != 1 {
		return Get{}, fmt.Errorf("expected 1 key entry to follow GET command but got %d entries: %v", len(data), data)
	}

	key, ok := data[0].(string)
	if !ok {
		return Get{}, fmt.Errorf("expected the input to the GET command to be a string key but it was %[1]v of type %[1]v", data[0])
	}

	return Get{key}, nil
}
