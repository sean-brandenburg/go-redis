package connection

import "net"

type ConnectionType string

const (
	ClientConnection  ConnectionType = "client"
	MasterConnection  ConnectionType = "master"
	ReplicaConnection ConnectionType = "replica"
)

type Connection interface {
	WriteString(string) (int, error)

	ReadNextCmdString() (string, error)

	ReadRDBFile() (string, error)

	ConnectionType() ConnectionType

	RemoteAddr() net.Addr

	LocalAddr() net.Addr

	Close() error
}
