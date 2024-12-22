package connection

import (
	"errors"
	"net"
	"time"
)

type ChannelConn struct {
	connType ConnectionType
	dataChan chan string
}

func NewChannelConn(connType ConnectionType) Connection {
	dataChan := make(chan string)
	return ChannelConn{
		dataChan: dataChan,
		connType: connType,
	}
}

func NewChannelConnWithBuffer(connType ConnectionType, bufferSize int) Connection {
	dataChan := make(chan string, bufferSize)
	return ChannelConn{
		dataChan: dataChan,
		connType: connType,
	}
}

func (p ChannelConn) WriteString(data string) (int, error) {
	select {
	case p.dataChan <- data:
	case <-time.After(time.Second):
		return 0, errors.New("failed to write to data channel before 1 second timeout")
	}
	return 0, nil
}

func (p ChannelConn) ReadNextCmdString() (string, error) {
	return p.readFromPipe()
}

func (p ChannelConn) ReadRDBFile() (string, error) {
	return p.readFromPipe()
}

func (p ChannelConn) readFromPipe() (string, error) {
	select {
	case data := <-p.dataChan:
		return data, nil
	case <-time.After(time.Second):
		return "", errors.New("failed to write to data channel before 1 second timeout")
	}
}

func (p ChannelConn) ConnectionType() ConnectionType {
	return p.connType
}

func (p ChannelConn) RemoteAddr() net.Addr {
	return nil
}

func (p ChannelConn) LocalAddr() net.Addr {
	return nil
}

func (p ChannelConn) Close() error {
	return nil
}
