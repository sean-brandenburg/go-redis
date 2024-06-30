package command

import (
	"fmt"
	"math"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/stretchr/testify/assert"
)

func TestParseSimpleString(t *testing.T) {
	for _, tc := range []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "+",
			expectedOutput: "",
		},
		{
			input:          "+0",
			expectedOutput: "0",
		},
		{
			input:          "+abc",
			expectedOutput: "abc",
		},
		{
			input:          "++++++",
			expectedOutput: "+++++",
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to string %q", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: []string{tc.input}}
			res, err := parser.parseSimpleString()
			assert.NoError(t, err)
			assert.Equal(t, res, tc.expectedOutput)
		})
	}
}

func TestParseInt(t *testing.T) {
	for _, tc := range []struct {
		input          string
		expectedOutput int64
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
			expectedOutput: math.MaxInt64,
		},
		{
			input:          fmt.Sprintf(":%d", math.MinInt),
			expectedOutput: math.MinInt64,
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to int %d", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: []string{tc.input}}
			res, err := parser.parseInt()
			assert.NoError(t, err)
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
			assert.NoError(t, err)
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
			expectedOutput: []any{int64(1)},
		},
		{
			input:          []string{"*2", "$4", "ECHO", "$4", "test"},
			expectedOutput: []any{"ECHO", "test"},
		},
		// Nested Array
		{
			input:          []string{"*2", "$1", "1", "*2", "$1", "2", "$1", "3"},
			expectedOutput: []any{"1", []any{"2", "3"}},
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to array %q", tc.input, tc.expectedOutput), func(t *testing.T) {
			parser := CommandParser{tokens: tc.input}
			res, err := parser.parseArray()
			assert.NoError(t, err)
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
			// NOTE: This test also checks for the case insensitivity of command parsing
			rawCmdString: "*1\r\n$4\r\nPiNg\r\n",
			expectedCmd:  Ping{},
		},
		{
			rawCmdString: "*2\r\n$4\r\nECHO\r\n$4\r\ntest\r\n",
			expectedCmd:  Echo{Payload: "test"},
		},
		{
			rawCmdString: "*3\r\n$3\r\nSET\r\n$6\r\nbanana\r\n$6\r\nyellow\r\n",
			expectedCmd:  Set{KeyPayload: "banana", ValuePayload: "yellow"},
		},
		{
			// NOTE: This also checks that px is case insensitive
			rawCmdString: "*5\r\n$3\r\nSET\r\n$6\r\nbanana\r\n$6\r\nyellow\r\n$2\r\npX\r\n$3\r\n100\r\n",
			expectedCmd:  Set{KeyPayload: "banana", ValuePayload: "yellow", ExpiryTimeMs: int64(100)},
		},
		{
			rawCmdString: "*2\r\n$3\r\nGET\r\n$6\r\nbanana\r\n",
			expectedCmd:  Get{Payload: "banana"},
		},
		{
			rawCmdString: "*2\r\n$4\r\nINFO\r\n$11\r\nreplication\r\n",
			expectedCmd:  Info{Payload: "replication"},
		},
	} {
		t.Run(fmt.Sprintf("input %q should parse to populated %T command", tc.rawCmdString, tc.expectedCmd), func(t *testing.T) {
			parser, err := NewParser(tc.rawCmdString, log.NewNoOpLogger())
			assert.NoError(t, err)

			cmd, err := parser.Parse()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCmd, cmd)
		})
	}
}
