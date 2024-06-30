package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEchoEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		echo              Echo
		expectedCmdString string
	}{
		{
			echo:              Echo{},
			expectedCmdString: "*2\r\n$4\r\necho\r\n$0\r\n\r\n",
		},
		{
			echo:              Echo{Payload: "test"},
			expectedCmdString: "*2\r\n$4\r\necho\r\n$4\r\ntest\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.echo.EncodedCommand()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}
