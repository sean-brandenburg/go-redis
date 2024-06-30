package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEncodeCommand(t *testing.T) {
	for _, tc := range []struct {
		get              Get
		expectedCmdString string
	}{
		{
			get:              Get{},
			expectedCmdString: "*2\r\n$3\r\nget\r\n$0\r\n\r\n",
		},
		{
			get:              Get{"test"},
			expectedCmdString: "*2\r\n$3\r\nget\r\n$4\r\ntest\r\n",
		},
	} {
		t.Run(fmt.Sprintf("should be able to encode command %q", tc.expectedCmdString), func(t *testing.T) {
			res, err := tc.get.EncodedCommand()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedCmdString, res)
		})
	}
}