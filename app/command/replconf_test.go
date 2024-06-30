package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplConfEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		conf              ReplConf
		expectedCmdString string
	}{
		{
			conf:              ReplConf{},
			expectedCmdString: "*3\r\n$8\r\nreplconf\r\n$0\r\n\r\n$0\r\n\r\n",
		},
		{
			conf:              ReplConf{KeyPayload: "key", ValuePayload: "val"},
			expectedCmdString: "*3\r\n$8\r\nreplconf\r\n$3\r\nkey\r\n$3\r\nval\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.conf.EncodedCommand()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}
