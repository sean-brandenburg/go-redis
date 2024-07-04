package server

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/stretchr/testify/assert"
)

func initTestCommandExecutor(initialData serverStore) (commandExecutor, net.Conn) {
	writer, reader := net.Pipe()
	return commandExecutor{
		server: &BaseServer{
			storeData:   initialData,
			storeDataMu: &sync.Mutex{},
			logger:      log.NewNoOpLogger(),
		},
		clientConn: writer,
	}, reader
}

func TestExecutePing(t *testing.T) {
	executor, reader := initTestCommandExecutor(map[string]storeValue{})

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Execute the command in a goroutine because reader.Write with a net.Pipe is blocking
	go func() {
		defer wg.Done()

		err := executor.executePing(command.Ping{})
		assert.Nil(t, err)
	}()

	res := make([]byte, MaxMessageSize)
	numBytes, err := reader.Read(res)
	wg.Wait()

	assert.Nil(t, err)
	assert.Equal(t, "+PONG\r\n", string(res[:numBytes]))
}

func TestExecuteEcho(t *testing.T) {
	for _, tc := range []struct {
		payload     string
		expectedRes string
	}{
		{
			payload:     "",
			expectedRes: "+\r\n",
		},
		{
			payload:     "a",
			expectedRes: "+a\r\n",
		},
		{
			payload:     "123",
			expectedRes: "+123\r\n",
		},
		{
			payload:     "[1,2,3]",
			expectedRes: "+[1,2,3]\r\n",
		},
		{
			payload:     "\r\n",
			expectedRes: "+\r\n\r\n",
		},
	} {
		t.Run(fmt.Sprintf("ECHO with value %q should return the encoded state", tc.payload), func(t *testing.T) {
			executor, reader := initTestCommandExecutor(serverStore{})

			wg := sync.WaitGroup{}
			wg.Add(1)
			// Execute the command in a goroutine because reader.Write with a net.Pipe is blocking
			go func() {
				err := executor.executeEcho(command.Echo{Payload: tc.payload})
				assert.Nil(t, err)
				wg.Done()
			}()

			res := make([]byte, MaxMessageSize)
			numBytes, err := reader.Read(res)
			wg.Wait()

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedRes, string(res[:numBytes]))
		})
	}
}

func TestExecuteGet(t *testing.T) {
	for _, tc := range []struct {
		inputKey                string
		initialServerStoreState serverStore
		expectedRes             string
	}{
		{
			inputKey:                "c",
			initialServerStoreState: serverStore{"a": {data: "b"}},
			expectedRes:             "$-1\r\n",
		},
		{
			inputKey:                "a",
			initialServerStoreState: serverStore{"a": {data: 2}},
			expectedRes:             ":2\r\n",
		},
		{
			inputKey:                "a",
			initialServerStoreState: serverStore{"a": {data: "b"}},
			expectedRes:             "+b\r\n",
		},
		{
			inputKey: "a",
			initialServerStoreState: serverStore{
				"a": {data: "b"},
				"c": {data: "d"},
			},
			expectedRes: "+b\r\n",
		},
	} {
		t.Run(fmt.Sprintf("GET with key %q and inital state %v and no expiry should succeed", tc.inputKey, tc.initialServerStoreState), func(t *testing.T) {
			executor, reader := initTestCommandExecutor(tc.initialServerStoreState)

			wg := sync.WaitGroup{}
			wg.Add(1)
			// Execute the command in a goroutine because reader.Write with a net.Pipe is blocking
			go func() {
				err := executor.executeGet(command.Get{Payload: tc.inputKey})
				assert.Nil(t, err)
				wg.Done()
			}()

			res := make([]byte, MaxMessageSize)
			numBytes, err := reader.Read(res)
			wg.Wait()

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedRes, string(res[:numBytes]))
		})
	}

	t.Run("GET on a key that has expired should delete it and return a null bulk string", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		executor, reader := initTestCommandExecutor(serverStore{
			"a": {data: "b", expiresAt: &pastTime},
		})

		wg := sync.WaitGroup{}
		wg.Add(1)
		// Execute the command in a goroutine because reader.Write with a net.Pipe is blocking
		go func() {
			err := executor.executeGet(command.Get{Payload: "a"})
			assert.Nil(t, err)
			wg.Done()
		}()

		res := make([]byte, MaxMessageSize)
		numBytes, err := reader.Read(res)
		wg.Wait()

		assert.Nil(t, err)
		assert.Zero(t, executor.server.Size())
		assert.Equal(t, command.NullBulkString, string(res[:numBytes]))
	})

	t.Run("GET on a key that has not expired should return it and should not modify the store state", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)
		executor, reader := initTestCommandExecutor(serverStore{
			"a": {data: "b", expiresAt: &futureTime},
		})

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := executor.executeGet(command.Get{Payload: "a"})
			assert.Nil(t, err)
			wg.Done()
		}()

		res := make([]byte, MaxMessageSize)
		numBytes, err := reader.Read(res)
		wg.Wait()

		assert.Nil(t, err)
		assert.Equal(t, "+b\r\n", string(res[:numBytes]))

		// Store data should not have been modified
		assert.Equal(t, 1, executor.server.Size())
		value, ok := executor.server.Get("a")
		assert.True(t, ok)
		assert.Equal(t, "b", value)
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
			executor, reader := initTestCommandExecutor(serverStore{})

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				err := executor.executeSet(command.Set{
					KeyPayload:   tc.inputKey,
					ValuePayload: tc.inputValue,
				})
				assert.Nil(t, err)
				wg.Done()
			}()

			res := make([]byte, MaxMessageSize)
			numBytes, err := reader.Read(res)
			wg.Wait()

			assert.Nil(t, err)
			assert.Equal(t, command.OKString, string(res[:numBytes]))
			assert.Equal(t, len(tc.expectedMapState), executor.server.Size())
			for expectedKey, expetedValue := range tc.expectedMapState {
				value, ok := executor.server.Get(expectedKey)
				assert.True(t, ok)
				assert.Equal(t, expetedValue.data, value)
			}
		})
	}

	t.Run("SET with an expiry time should create an unexpired object", func(t *testing.T) {
		executor, reader := initTestCommandExecutor(serverStore{})

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := executor.executeSet(command.Set{
				KeyPayload:   "a",
				ValuePayload: "b",
				ExpiryTimeMs: 10000,
			})
			assert.Nil(t, err)
			wg.Done()
		}()

		res := make([]byte, MaxMessageSize)
		numBytes, err := reader.Read(res)
		wg.Wait()

		assert.Nil(t, err)
		assert.Equal(t, command.OKString, string(res[:numBytes]))
		assert.Equal(t, 1, executor.server.Size())

		value, ok := executor.server.Get("a")
		assert.True(t, ok)
		assert.Equal(t, "b", value)
	})
}

// TODO: Fix these once info has a stable return value
// func TestExecuteInfo(t *testing.T) {
// 	for _, tc := range []struct {
// 		inputServer  Server
// 		expectedInfo string
// 	}{
// 		{
// 			inputServer: &BaseServer{
// 				logger: log.NewNoOpLogger(),
// 			},
// 			expectedInfo: "",
// 		},
// 		{
// 			inputServer: &MasterServer{
// 				BaseServer: BaseServer{
// 					logger: log.NewNoOpLogger(),
// 				},
// 			},
// 			expectedInfo: "",
// 		},
// 		{
// 			inputServer: &SlaveServer{
// 				BaseServer: BaseServer{
// 					logger: log.NewNoOpLogger(),
// 				},
// 				masterAddress: "localhost:123",
// 			},
// 			expectedInfo: "",
// 		},
// 	} {
// 		t.Run(fmt.Sprintf("executing INFO on a server of type %T should return the expected value", tc.inputServer), func(t *testing.T) {
// 			res, err := executeInfo(tc.inputServer, command.Info{
// 				Payload: "replication",
// 			})
// 			assert.Nil(t, err)
// 			assert.Equal(t, tc.expectedInfo, res)
// 		})
// 	}
// }

func TestExecuteCommand(t *testing.T) {
	for _, tc := range []struct {
		inputCommand command.Command
		expectedRes  string
	}{
		{
			inputCommand: command.Ping{},
			expectedRes:  "+PONG\r\n",
		},
		// TODO: Fix these once info has a stable return value
		// {
		// 	inputCommand: command.Info{Payload: "replication"},
		// 	expectedRes:  command.MustEncode("info res"),
		// },
		{
			inputCommand: command.Echo{Payload: "test"},
			expectedRes:  "+test\r\n",
		},
		{
			inputCommand: command.Get{Payload: "a"},
			expectedRes:  "+b\r\n",
		},
		{
			inputCommand: command.Set{KeyPayload: "c", ValuePayload: "d"},
			expectedRes:  "+OK\r\n",
		},
		{
			inputCommand: command.ReplConf{KeyPayload: "c", ValuePayload: "d"},
			expectedRes:  "+OK\r\n",
		},
		{
			inputCommand: command.PSync{ReplicationID: command.HARDCODE_REPL_ID, MasterOffset: "0"},
			expectedRes:  "+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n",
		},
		{
			inputCommand: command.Info{Payload: "replication"},
			expectedRes:  "$86\r\nmaster_repl_offset:0\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nrole:base\n\r\n",
		},
	} {
		t.Run(fmt.Sprintf("executing command %q should succeed", tc.inputCommand.String()), func(t *testing.T) {
			executor, reader := initTestCommandExecutor(serverStore{"a": {data: "b"}})

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				err := executor.execute(tc.inputCommand)
				assert.Nil(t, err)
				wg.Done()
			}()

			res := make([]byte, MaxMessageSize)
			numBytes, err := reader.Read(res)
			wg.Wait()

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedRes, string(res[:numBytes]))
		})
	}
}
