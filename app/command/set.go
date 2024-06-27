package command

import (
	"fmt"
)

type Set struct {
	KeyPayload   string
	ValuePayload any
}

func (set Set) String() string {
	return fmt.Sprintf("SET: %q -> %v", set.KeyPayload, set.ValuePayload)
}

func toSet(data []any) (Set, error) {
	if len(data) != 2 {
		return Set{}, fmt.Errorf("expected 1 key and 1 value entry to follow SET command but got %d entries: %v", len(data), data)
	}

	key, ok := data[0].(string)
	if !ok {
		return Set{}, fmt.Errorf("expected the first element in the SET command to be a string key but it was %[1]v of type %[1]v", data[0])
	}

	return Set{key, data[1]}, nil
}
