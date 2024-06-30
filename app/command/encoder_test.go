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
			expectedOutput: "#t\r\n",
		},
		{
			input:          false,
			expectedOutput: "#f\r\n",
		},
	} {
		t.Run(fmt.Sprintf("encoding input %v", tc.input), func(t *testing.T) {
			e := Encoder{}
			res, err := e.Encode(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, res)
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	for _, tc := range []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "string",
			expectedOutput: "$6\r\nstring",
		},
		{
			input:          "",
			expectedOutput: "$0\r\n",
		},
		{
			input:          "test\ntest",
			expectedOutput: "$9\r\ntest\ntest",
		},
	} {
		t.Run(fmt.Sprintf("encoding input %v", tc.input), func(t *testing.T) {
			res, err := encodeBulkString(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, res)
		})
	}
}
