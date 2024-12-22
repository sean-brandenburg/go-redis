package connection

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/log"
)

func NewLoggerNoopConn(l log.Logger, connType ConnectionType) Connection {
	return LogNoopConn{
		Logger:   l,
		ConnType: connType,
	}
}

type LogNoopConn struct {
	Logger   log.Logger
	ConnType ConnectionType
}

func (n LogNoopConn) WriteString(data string) (int, error) {
	n.Logger.Info("log noop conn WriteString() called")
	return 0, nil
}

func (n LogNoopConn) ReadNextCmdString() (string, error) {
	n.Logger.Info("log noop conn ReadNextCmdString() called")
	return "", nil
}

func (n LogNoopConn) ReadRDBFile() (string, error) {
	n.Logger.Info("log noop conn ReadRDBFile() called")
	return "", nil
}

func (n LogNoopConn) Close() error {
	n.Logger.Info("log noop conn Close() called")
	return nil
}

func (n LogNoopConn) LocalAddr() net.Addr {
	n.Logger.Info("log noop conn LocalAddr() called")
	return nil
}

func (n LogNoopConn) RemoteAddr() net.Addr {
	n.Logger.Info("log noop conn RemoteAddr() called")
	return nil
}

func (n LogNoopConn) ConnectionType() ConnectionType {
	n.Logger.Info("log noop conn ConnectionType() called")
	return n.ConnType
}
