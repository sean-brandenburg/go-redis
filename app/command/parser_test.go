package command

import (
	"fmt"
	"testing"
)

func TestParseInt(t *testing.T) {
	for _, tc := range []struct {
		input          string
		expectedOutput int
	}{
		{
			input:          ":0",
			expectedOutput: 0,
		},
		{
			input:          ":1",
			expectedOutput: 1,
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to %d", tc.input, tc.expectedOutput), func(t *testing.T) {
		})
	}
}
