package server

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/connection"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

func getTestMasterServer(initialData serverStore) Server {
	return &MasterServer{
		BaseServer: BaseServer{
			storeData:   initialData,
			storeDataMu: &sync.Mutex{},
			logger:      log.NewNoOpLogger(),
		},
		registeredReplicaConns: []connection.Connection{},
	}
}

func getTestReplicaServer(initialData serverStore) Server {
	return &ReplicaServer{
		BaseServer: BaseServer{
			storeData:   initialData,
			storeDataMu: &sync.Mutex{},
			logger:      log.NewNoOpLogger(),
		},
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

	conn := connection.NewChannelConn(connection.ClientConnection)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		err := RunCommand(srv, conn, cmd)
		assert.Nil(t, err)
		wg.Done()
	}()

	for idx, expectedMessage := range expectedOutputs {
		errContextMsg := fmt.Sprintf("unexpected message received on message number %d", idx)
		msg, err := conn.ReadNextCmdString()

		assert.Nil(t, err, errContextMsg)
		assert.Equal(t, expectedMessage, msg, errContextMsg)
	}

	// Wait group enforces that run command test failure will happen during the test rather than after
	// it finishes which gives us a cleaner error message
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

func TestExecuteInfo(t *testing.T) {
	runCommandAndCheckOutput(
		t,
		command.Info{Payload: "replication"},
		"$88\r\nmaster_repl_offset:0\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nrole:master\n\r\n",
	)
}

func TestExecuteWait(t *testing.T) {
	for _, tc := range []struct {
		numReplicas int64
		waitTimeMs  int64
	}{
		{
			numReplicas: 10,
			waitTimeMs:  20,
		},

		{
			numReplicas: 0,
			waitTimeMs:  0,
		},
	} {
		runCommandAndCheckOutput(t, command.Wait{NumReplicas: tc.numReplicas, WaitForMs: tc.waitTimeMs}, ":0\r\n")
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
		server      Server
		expectedRes string
	}{
		{
			key:         "capa",
			value:       "psync2",
			server:      getTestMasterServer(serverStore{}),
			expectedRes: command.OKString,
		},
		{
			key:         "GetAck",
			value:       "*",
			server:      getTestReplicaServer(serverStore{}),
			expectedRes: "*3\r\n+REPLCONF\r\n+ACK\r\n+0\r\n",
		},
	} {
		t.Run(fmt.Sprintf("PSYNC with key %q and value %q should return the expected res", tc.key, tc.value), func(t *testing.T) {
			runCommandAndCheckOutputWithServer(t, tc.server, command.ReplConf{Payload: []string{tc.key, tc.value}}, tc.expectedRes)
		})
	}
}

func TestExecutePSync(t *testing.T) {
	runCommandAndCheckOutputs(
		t,
		command.PSync{ReplicationID: command.HARDCODE_REPL_ID, MasterOffset: "0"},
		[]string{"+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n", fmt.Sprintf("$88\r\n%s", command.GetHardedCodedEmptyRDBBytes())},
	)
}
