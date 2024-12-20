package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		cmd               Command
		expectedCmdString string
	}{

		{
			cmd:               Ping{},
			expectedCmdString: "*1\r\n$4\r\nping\r\n",
		},
		{
			cmd:               Info{},
			expectedCmdString: "*2\r\n$4\r\ninfo\r\n$0\r\n\r\n",
		},
		{
			cmd:               Info{"test"},
			expectedCmdString: "*2\r\n$4\r\ninfo\r\n$4\r\ntest\r\n",
		},
		{
			cmd:               Echo{},
			expectedCmdString: "*2\r\n$4\r\necho\r\n$0\r\n\r\n",
		},
		{
			cmd:               Echo{Payload: "test"},
			expectedCmdString: "*2\r\n$4\r\necho\r\n$4\r\ntest\r\n",
		},
		{
			cmd:               Get{},
			expectedCmdString: "*2\r\n$3\r\nget\r\n$0\r\n\r\n",
		},
		{
			cmd:               Get{"test"},
			expectedCmdString: "*2\r\n$3\r\nget\r\n$4\r\ntest\r\n",
		},
		{
			cmd:               ReplConf{},
			expectedCmdString: "*3\r\n$8\r\nreplconf\r\n$0\r\n\r\n$0\r\n\r\n",
		},
		{
			cmd:               ReplConf{Payload: []string{"key", "val"}},
			expectedCmdString: "*3\r\n$8\r\nreplconf\r\n$3\r\nkey\r\n$3\r\nval\r\n",
		},
		{
			cmd:               Set{KeyPayload: "key", ValuePayload: "val"},
			expectedCmdString: "*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n",
		},
		{
			cmd:               Set{KeyPayload: "key", ValuePayload: "val", ExpiryTimeMs: 100},
			expectedCmdString: "*5\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n$2\r\npx\r\n$3\r\n100\r\n",
		},
		{
			cmd:               PSync{},
			expectedCmdString: "*3\r\n$5\r\npsync\r\n$0\r\n\r\n$0\r\n\r\n",
		},
		{
			cmd:               PSync{ReplicationID: "2", MasterOffset: "1"},
			expectedCmdString: "*3\r\n$5\r\npsync\r\n$1\r\n2\r\n$1\r\n1\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.cmd.EncodedCommand()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}
