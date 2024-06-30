package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingEncodeCommand(t *testing.T) {
	res, err := Ping{}.EncodedCommand()
	assert.Nil(t, err)
	assert.Equal(t, "*1\r\n$4\r\nping\r\n", res)
}
