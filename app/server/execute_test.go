package server

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

func initTestCommandExecutor(initialData serverStore) (commandExecutor, net.Conn) {
	writer, reader := net.Pipe()
	return commandExecutor{
		server: &MasterServer{
			BaseServer: BaseServer{
				storeData:   initialData,
				storeDataMu: &sync.Mutex{},
				logger:      log.NewNoOpLogger(),
			},
			registeredReplicaConns: []net.Conn{},
		},
		conn: ConnWithType{
			Conn:     writer,
			ConnType: ClientConnection,
		},
	}, reader
}

func getTestMasterServer(initialData serverStore) Server {
	return &MasterServer{
		BaseServer: BaseServer{
			storeData:   initialData,
			storeDataMu: &sync.Mutex{},
			logger:      log.NewNoOpLogger(),
		},
		registeredReplicaConns: []net.Conn{},
	}
}

func runCommandAndCheckOutput(t *testing.T, cmd command.Command, expectedOutput string) {
	runCommandAndCheckOutputWithServer(t, getTestMasterServer(serverStore{}), cmd, expectedOutput)
}

func runCommandAndCheckOutputs(t *testing.T, cmd command.Command, expectedOutputs []string) {
	runCommandAndCheckOutputsWithServer(t, getTestMasterServer(serverStore{}), cmd, expectedOutputs)
}

func runCommandAndCheckOutputWithServer(t *testing.T, srv Server, cmd command.Command, expectedOutput string) {
	runCommandAndCheckOutputsWithServer(t, srv, cmd, []string{expectedOutput})
}

// Sends a command to the specified server and checks that it responds with the expectedOutputs
func runCommandAndCheckOutputsWithServer(t *testing.T, srv Server, cmd command.Command, expectedOutputs []string) {
	t.Helper()

	writer, reader := net.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := RunCommand(
			srv,
			ConnWithType{
				Conn:     writer,
				ConnType: ClientConnection,
			},
			cmd,
		)
		assert.Nil(t, err)
	}()

	for idx, expectedMessage := range expectedOutputs {
		errContextMsg := fmt.Sprintf("unexpected message received on message number %d", idx)
		res := make([]byte, MaxMessageSize)
		numBytes, err := reader.Read(res)

		assert.Nil(t, err, errContextMsg)
		assert.Equal(t, expectedMessage, string(res[:numBytes]), errContextMsg)
	}

	wg.Wait()
}

func TestExecutePing(t *testing.T) {
	runCommandAndCheckOutput(t, command.Ping{}, "+PONG\r\n")
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
			runCommandAndCheckOutput(t, command.Echo{Payload: tc.payload}, tc.expectedRes)
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
			server := getTestMasterServer(tc.initialServerStoreState)
			runCommandAndCheckOutputWithServer(t, server, command.Get{Payload: tc.inputKey}, tc.expectedRes)
		})
	}

	t.Run("GET on a key that has expired should delete it and return a null bulk string", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		server := getTestMasterServer(serverStore{"a": {data: "b", expiresAt: &pastTime}})
		runCommandAndCheckOutputWithServer(t, server, command.Get{Payload: "a"}, command.NullBulkString)
		_, ok := server.Get("a")
		assert.False(t, ok)
	})

	t.Run("GET on a key that has not expired should return it and should not modify the store state", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)
		server := getTestMasterServer(serverStore{"a": {data: "b", expiresAt: &futureTime}})
		runCommandAndCheckOutputWithServer(t, server, command.Get{Payload: "a"}, "+b\r\n")
		value, ok := server.Get("a")
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
			server := getTestMasterServer(tc.initialMapState)
			runCommandAndCheckOutputWithServer(t, server, command.Set{KeyPayload: tc.inputKey, ValuePayload: tc.inputValue}, command.OKString)

			assert.Equal(t, len(tc.expectedMapState), server.Size())
			for expectedKey, expetedValue := range tc.expectedMapState {
				value, ok := server.Get(expectedKey)
				assert.True(t, ok)
				assert.Equal(t, expetedValue.data, value)
			}
		})
	}

	t.Run("SET with an expiry time should create an unexpired object", func(t *testing.T) {
		server := getTestMasterServer(serverStore{})
		runCommandAndCheckOutputWithServer(t, server, command.Set{
			KeyPayload:   "a",
			ValuePayload: "b",
			ExpiryTimeMs: 10000,
		}, command.OKString)

		assert.Equal(t, 1, server.Size())

		value, ok := server.Get("a")
		assert.True(t, ok)
		assert.Equal(t, "b", value)
	})
}

func TestExecuteReplConf(t *testing.T) {
	for _, tc := range []struct {
		key         string
		value       string
		expectedRes string
	}{
		{
			key:         "c",
			value:       "d",
			expectedRes: command.OKString,
		},

		{
			key:         "GetAck",
			value:       "*",
			expectedRes: "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n",
		},
	} {
		t.Run(fmt.Sprintf("PSYNC with key %q and value %q should return the expected res", tc.key, tc.value), func(t *testing.T) {
			runCommandAndCheckOutput(t, command.ReplConf{KeyPayload: tc.key, ValuePayload: tc.value}, tc.expectedRes)
		})
	}
}

func TestExecuteInfo(t *testing.T) {
	runCommandAndCheckOutput(
		t,
		command.Info{Payload: "replication"},
		"$88\r\nmaster_repl_offset:0\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nrole:master\n\r\n",
	)
}

func TestExecutePSync(t *testing.T) {
	runCommandAndCheckOutputs(
		t,
		command.PSync{ReplicationID: command.HARDCODE_REPL_ID, MasterOffset: "0"},
		[]string{"+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n", fmt.Sprintf("$88\r\n%s", command.GetHardedCodedEmptyRDBBytes())},
	)
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
// 			inputServer: &ReplicaServer{
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

// TODO: Looped tests for replica vs master with expected behavior
