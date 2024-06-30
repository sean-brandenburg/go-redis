package server

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/stretchr/testify/assert"
)

func TestExecutePing(t *testing.T) {
	res, err := executePing(command.Ping{})
	assert.Nil(t, err)
	assert.Equal(t, command.MustEncode("PONG"), res)
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
		t.Run(fmt.Sprintf("ECHO with value %q should return the encoded state", tc.payload), func(t *testing.T) {
			res, err := executeEcho(command.Echo{Payload: tc.payload})
			assert.Nil(t, err)
			assert.Equal(t, command.MustEncode(tc.payload), res)
		})
	}
}

func TestExecuteGet(t *testing.T) {
	for _, tc := range []struct {
		inputKey       string
		mapState       map[string]storeValue
		expectedResult string
	}{
		{
			inputKey:       "c",
			mapState:       map[string]storeValue{"a": {data: "b"}},
			expectedResult: "$-1\r\n",
		},
		{
			inputKey:       "a",
			mapState:       map[string]storeValue{"a": {data: 2}},
			expectedResult: ":2\r\n",
		},
		{
			inputKey:       "a",
			mapState:       map[string]storeValue{"a": {data: "b"}},
			expectedResult: "+b\r\n",
		},
		{
			inputKey: "a",
			mapState: map[string]storeValue{
				"a": {data: "b"},
				"c": {data: "d"},
			},
			expectedResult: "+b\r\n",
		},
	} {
		t.Run(fmt.Sprintf("GET with key %q and inital state %v and no expiry should succeed", tc.inputKey, tc.mapState), func(t *testing.T) {
			server := BaseServer{
				logger:      log.NewNoOpLogger(),
				storeData:   tc.mapState,
				storeDataMu: &sync.Mutex{},
			}

			res, err := executeGet(&server, command.Get{Payload: tc.inputKey})
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}

	t.Run("GET on a key that has expired should delete it and return a null bulk string", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		server := BaseServer{
			logger: log.NewNoOpLogger(),
			storeData: map[string]storeValue{
				"a": {data: "b", expiresAt: &pastTime},
			},
			storeDataMu: &sync.Mutex{},
		}

		res, err := executeGet(&server, command.Get{Payload: "a"})
		assert.Nil(t, err)
		assert.Equal(t, res, command.NullBulkString)
		assert.Empty(t, server.storeData)
	})

	t.Run("GET on a key that has not expired should return it and should not modify the store state", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)

		server := BaseServer{
			logger: log.NewNoOpLogger(),
			storeData: map[string]storeValue{
				"a": {data: "b", expiresAt: &futureTime},
			},
			storeDataMu: &sync.Mutex{},
		}

		res, err := executeGet(&server, command.Get{Payload: "a"})
		assert.Nil(t, err)
		assert.Equal(t, res, command.MustEncode("b"))
		// Store data should not have been modified
		assert.Equal(
			t,
			map[string]storeValue{"a": {data: "b", expiresAt: &futureTime}},
			server.storeData,
		)
	})
}

func TestExecuteSet(t *testing.T) {
	for _, tc := range []struct {
		inputKey         string
		inputValue       any
		initialMapState  map[string]storeValue
		expectedMapState map[string]storeValue
	}{
		{
			inputKey:         "a",
			inputValue:       1,
			initialMapState:  map[string]storeValue{},
			expectedMapState: map[string]storeValue{"a": {data: 1}},
		},
		{
			inputKey:         "a",
			inputValue:       1,
			initialMapState:  map[string]storeValue{"a": {data: 2}},
			expectedMapState: map[string]storeValue{"a": {data: 1}},
		},
		{
			inputKey:         "a",
			inputValue:       "b",
			initialMapState:  map[string]storeValue{"a": {data: 1}},
			expectedMapState: map[string]storeValue{"a": {data: "b"}},
		},
	} {
		t.Run(fmt.Sprintf("SET with key %q and value %v should properly update the server state", tc.inputKey, tc.inputValue), func(t *testing.T) {
			server := BaseServer{
				logger:      log.NewNoOpLogger(),
				storeData:   tc.initialMapState,
				storeDataMu: &sync.Mutex{},
			}

			res, err := executeSet(&server, command.Set{
				KeyPayload:   tc.inputKey,
				ValuePayload: tc.inputValue,
			})
			assert.Nil(t, err)
			assert.Equal(t, command.OKString, res)

			assert.Equal(t, tc.expectedMapState, server.storeData)
		})
	}

	t.Run("SET with an expiry time should set the expiry date to a time in the future", func(t *testing.T) {
		server := BaseServer{
			logger:      log.NewNoOpLogger(),
			storeData:   make(map[string]storeValue, 1),
			storeDataMu: &sync.Mutex{},
		}

		res, err := executeSet(&server, command.Set{
			KeyPayload:   "a",
			ValuePayload: "b",
			ExpiryTimeMs: 10000,
		})
		assert.Nil(t, err)
		assert.Equal(t, command.OKString, res)

		// Expiry time should be in the future
		assert.True(t, server.storeData["a"].expiresAt.After(time.Now()))
	})
}

// TODO: Fix these once info has a stable return value
func TestExecuteInfo(t *testing.T) {
	for _, tc := range []struct {
		inputServer  Server
		expectedInfo string
	}{
		{
			inputServer: &BaseServer{
				logger: log.NewNoOpLogger(),
			},
			expectedInfo: "",
		},
		{
			inputServer: &MasterServer{
				BaseServer: BaseServer{
					logger: log.NewNoOpLogger(),
				},
			},
			expectedInfo: "",
		},
		{
			inputServer: &SlaveServer{
				BaseServer: BaseServer{
					logger: log.NewNoOpLogger(),
				},
				masterAddress: "localhost:123",
			},
			expectedInfo: "",
		},
	} {
		t.Run(fmt.Sprintf("executing INFO on a server of type %T should return the expected value", tc.inputServer), func(t *testing.T) {
			res, err := executeInfo(tc.inputServer, command.Info{
				Payload: "replication",
			})
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedInfo, res)
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	for _, tc := range []struct {
		inputCommand command.Command
		expectedRes  string
	}{
		{
			inputCommand: command.Ping{},
			expectedRes:  command.MustEncode("PONG"),
		},
		// TODO: Fix these once info has a stable return value
		// {
		// 	inputCommand: command.Info{Payload: "replication"},
		// 	expectedRes:  command.MustEncode("info res"),
		// },
		{
			inputCommand: command.Echo{Payload: "test"},
			expectedRes:  command.MustEncode("test"),
		},
		{
			inputCommand: command.Get{Payload: "a"},
			expectedRes:  command.MustEncode("b"),
		},
		{
			inputCommand: command.Set{KeyPayload: "c", ValuePayload: "d"},
			expectedRes:  command.MustEncode("OK"),
		},
	} {
		t.Run(fmt.Sprintf("executing command %q should succeed", tc.inputCommand.String()), func(t *testing.T) {
			server := BaseServer{
				logger:      log.NewNoOpLogger(),
				storeData:   map[string]storeValue{"a": {data: "b"}},
				storeDataMu: &sync.Mutex{},
			}

			res, err := executeCommand(&server, tc.inputCommand)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedRes, res)
		})
	}
}
