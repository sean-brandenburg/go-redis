package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	for _, tc := range []struct {
		input          any
		expectedOutput string
	}{
		{
			input:          "string",
			expectedOutput: "+string\r\n",
		},
		{
			input:          100,
			expectedOutput: ":100\r\n",
		},
		{
			input:          []any{1, 2, "3"},
			expectedOutput: "*3\r\n:1\r\n:2\r\n+3\r\n",
		},
		{
			input:          true,
			expectedOutput: "#true\r\n",
		},
		{
			input:          false,
			expectedOutput: "#false\r\n",
		},
	} {
		t.Run(fmt.Sprintf("encoding input %v", tc.input), func(t *testing.T) {
			res, err := Encode(tc.input)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, res)
		})
	}
}
