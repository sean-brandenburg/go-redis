package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfoEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		info              Info
		expectedCmdString string
	}{
		{
			info:              Info{},
			expectedCmdString: "*2\r\n$4\r\ninfo\r\n$0\r\n\r\n",
		},
		{
			info:              Info{"test"},
			expectedCmdString: "*2\r\n$4\r\ninfo\r\n$4\r\ntest\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.info.EncodedCommand()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}
