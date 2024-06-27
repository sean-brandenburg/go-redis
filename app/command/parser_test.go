package command

import (
	"fmt"
	"math"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/stretchr/testify/assert"
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
		{
			input:          ":-1",
			expectedOutput: -1,
		},
		{
			input:          fmt.Sprintf(":%d", math.MaxInt),
			expectedOutput: math.MaxInt,
		},
		{
			input:          fmt.Sprintf(":%d", math.MinInt),
			expectedOutput: math.MinInt,
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to int %d", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: []string{tc.input}}
			res, err := parser.parseInt()
			assert.Nil(t, err)
			assert.Equal(t, res, tc.expectedOutput)
		})
	}
}

func TestParseBulkString(t *testing.T) {
	for _, tc := range []struct {
		input          []string
		expectedOutput string
	}{
		{
			input:          []string{"$0", ""},
			expectedOutput: "",
		},
		{
			input:          []string{"$1", "a"},
			expectedOutput: "a",
		},
		{
			input:          []string{"$3", "xyz"},
			expectedOutput: "xyz",
		},
		{
			// This is an edge case where the input string contains the delimeter \r\n
			input:          []string{"$6", "abc", "d"},
			expectedOutput: "abc\r\nd",
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to string %q", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: tc.input}
			res, err := parser.parseBulkString()
			assert.Nil(t, err)
			assert.Equal(t, res, tc.expectedOutput)
		})
	}
}

func TestParseArray(t *testing.T) {
	for _, tc := range []struct {
		input          []string
		expectedOutput []any
	}{
		{
			// Empty Array: "*0\r\n"
			input:          []string{"*0"},
			expectedOutput: []any{},
		},
		{
			input:          []string{"*1", ":1"},
			expectedOutput: []any{1},
		},
		{
			input:          []string{"*2", "$4", "ECHO", "$4", "test"},
			expectedOutput: []any{"ECHO", "test"},
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to array %q", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: tc.input}
			res, err := parser.parseArray()
			assert.Nil(t, err)
			assert.Equal(t, res, tc.expectedOutput)
		})
	}
}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		rawCmdString string
		expectedCmd  Command
	}{
		{
			rawCmdString: "*2\r\n$4\r\nECHO\r\n$4\r\ntest\r\n",
			expectedCmd:  Echo{Payload: "test"},
		},
		{
			rawCmdString: "*1\r\n$4\r\nPING\r\n",
			expectedCmd:  Ping{},
		},
		{
			rawCmdString: "*3\r\n$3\r\nSET\r\n$6\r\nbanana\r\n$6\r\nyellow\r\n",
			expectedCmd:  Set{KeyPayload: "banana", ValuePayload: "yellow"},
		},
		{
			rawCmdString: "*2\r\n$3\r\nGET\r\n$6\r\nbanana\r\n",
			expectedCmd:  Get{Payload: "banana"},
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to populated %T command", tc.rawCmdString, tc.expectedCmd), func(t *testing.T) {
			parser, err := NewParser(tc.rawCmdString, log.NewNoOpLogger())
			assert.Nil(t, err)

			cmd, err := parser.Parse()
			assert.Nil(t, err)
			assert.Equal(t, cmd, tc.expectedCmd)
		})
	}
}
