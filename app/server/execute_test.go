package server

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/stretchr/testify/assert"
)

func TestExecutePing(t *testing.T) {
	server := Server{
		Logger: log.NewNoOpLogger(),
	}

	res, err := server.executePing(command.Ping{})
	assert.Nil(t, err)
	assert.Equal(t, "+PONG\r\n", res)
}

func TestExecuteEcho(t *testing.T) {
	for _, tc := range []struct {
		payload string
	}{
		{payload: ""},
		{payload: "a"},
		{payload: "123"},
		{payload: "[1,2,3]"},
		{payload: "\r\n"},
	} {
		t.Run(fmt.Sprintf("echo with value %q", tc.payload), func(t *testing.T) {
			server := Server{
				Logger: log.NewNoOpLogger(),
			}

			res, err := server.executeEcho(command.Echo{Payload: tc.payload})
			assert.Nil(t, err)
			assert.Equal(t, fmt.Sprintf("+%s\r\n", tc.payload), res)
		})
	}
}

func TestExecuteGet(t *testing.T) {
	for _, tc := range []struct {
		inputKey       string
		mapState       map[string]any
		expectedResult string
	}{
		{
			inputKey:       "a",
			mapState:       map[string]any{"a": 2},
			expectedResult: ":2\r\n",
		},
		{
			inputKey:       "a",
			mapState:       map[string]any{"a": "b"},
			expectedResult: "+b\r\n",
		},
		{
			inputKey: "a",
			mapState: map[string]any{
				"a": "b",
				"c": "d",
			},
			expectedResult: "+b\r\n",
		},
	} {
		t.Run("", func(t *testing.T) {
			server := Server{
				Logger:    log.NewNoOpLogger(),
				StoreData: tc.mapState,
			}

			res, err := server.executeGet(command.Get{Payload: tc.inputKey})
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}

func TestExecuteSet(t *testing.T) {
	for _, tc := range []struct {
		inputKey         string
		inputValue       any
		initialMapState  map[string]any
		expectedMapState map[string]any
	}{
		{
			inputKey:         "a",
			inputValue:       1,
			initialMapState:  map[string]any{},
			expectedMapState: map[string]any{"a": 1},
		},
		{
			inputKey:         "a",
			inputValue:       1,
			initialMapState:  map[string]any{"a": 2},
			expectedMapState: map[string]any{"a": 1},
		},
		{
			inputKey:         "a",
			inputValue:       "b",
			initialMapState:  map[string]any{"a": 1},
			expectedMapState: map[string]any{"a": "b"},
		},
	} {
		t.Run("", func(t *testing.T) {
			server := Server{
				Logger:    log.NewNoOpLogger(),
				StoreData: tc.initialMapState,
			}

			res, err := server.executeSet(command.Set{KeyPayload: tc.inputKey, ValuePayload: tc.inputValue})
			assert.Nil(t, err)
			assert.Equal(t, "+OK\r\n", res)

			assert.Equal(t, tc.expectedMapState, server.StoreData)
		})
	}
}
