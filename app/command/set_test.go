package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		set               Set
		expectedCmdString string
	}{
		{
			set:               Set{KeyPayload: "key", ValuePayload: "val"},
			expectedCmdString: "*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n",
		},
		{
			set:               Set{KeyPayload: "key", ValuePayload: "val", ExpiryTimeMs: 100},
			expectedCmdString: "*5\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n$2\r\npx\r\n$3\r\n100\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.set.EncodedCommand()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}
