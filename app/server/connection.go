package server

import (
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type ConnectionType string

const (
	ClientConnection  ConnectionType = "client"
	MasterConnection  ConnectionType = "master"
	ReplicaConnection ConnectionType = "replica"
)

type Connection interface {
	net.Conn

	ConnectionType() ConnectionType
}

type ConnWithType struct {
	net.Conn

	ConnType ConnectionType
}

func (c ConnWithType) ConnectionType() ConnectionType {
	return c.ConnType
}

type LogNoopConn struct {
	Logger   log.Logger
	ConnType ConnectionType
}

func (n LogNoopConn) Read(_ []byte) (int, error) {
	n.Logger.Info("log noop conn Read() called")
	return 0, nil
}

func (n LogNoopConn) Write(_ []byte) (int, error) {
	n.Logger.Info("log noop conn Write() called")
	return 0, nil
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

func (n LogNoopConn) SetDeadline(_ time.Time) error {
	n.Logger.Info("log noop conn SetDeadline() called")
	return nil
}

func (n LogNoopConn) SetReadDeadline(_ time.Time) error {
	n.Logger.Info("log noop conn SetReadDeadline() called")
	return nil
}

func (n LogNoopConn) SetWriteDeadline(_ time.Time) error {
	n.Logger.Info("log noop conn SetWriteDeadline() called")
	return nil
}

func (n LogNoopConn) ConnectionType() ConnectionType {
	n.Logger.Info("log noop conn ConnectionType() called")
	return n.ConnType
}
