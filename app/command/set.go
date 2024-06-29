package command

import (
	"fmt"
	"strconv"
	"strings"
)

type Set struct {
	KeyPayload   string
	ValuePayload any

	// Set a lifetime for the existence of this key value
	ExpiryTimeMs int64
}

func (set Set) String() string {
	return fmt.Sprintf("SET: (%q -> %v) with expiration %d", set.KeyPayload, set.ValuePayload, set.ExpiryTimeMs)
}

// TODO: Should clean up this function
func toSet(data []any) (Set, error) {
	if len(data) != 2 && len(data) != 4 {
		return Set{}, fmt.Errorf("expected 1 key and 1 value entry to follow SET command but got %d entries: %v", len(data), data)
	}

	rawKey := data[0]
	key, ok := rawKey.(string)
	if !ok {
		return Set{}, fmt.Errorf("expected the first element in the SET command to be a string key but it was %[1]v of type %[1]v", rawKey)
	}

	// TODO: If there are more of these flags, I should make a better system for handling these
	// For now just hard code a check for the px flag
	timeout := int64(0)
	if len(data) == 4 {
		rawFlag := data[2]
		rawFlagValue := data[3]

		if flag, ok := rawFlag.(string); !ok || strings.ToLower(flag) != "px" {
			return Set{}, fmt.Errorf("received invalid parameter for SET operation. flag %[1]v of type %[1]v is not a known option", rawFlag)
		}
		rawTimeoutStr, ok := rawFlagValue.(string)
		if !ok {
			return Set{}, fmt.Errorf("expected the value of the set PX option but it was %[1]v of type %[1]v", rawFlagValue)
		}

		var err error
		timeout, err = strconv.ParseInt(rawTimeoutStr, 10, 64)
		if err != nil {
			return Set{}, fmt.Errorf("expected the value of the SET - PX option to be an int64 but it was %q", rawTimeoutStr)
		}
	}

	return Set{
		KeyPayload:   key,
		ValuePayload: data[1],
		ExpiryTimeMs: timeout,
	}, nil
}
