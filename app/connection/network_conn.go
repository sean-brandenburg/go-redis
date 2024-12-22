package connection

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/log"
)

type NetworkConn struct {
	readWriter *bufio.ReadWriter

	conn net.Conn

	connType ConnectionType

	logger log.Logger
}

func NewNetworkConn(conn net.Conn, connType ConnectionType, logger log.Logger) Connection {
	return NetworkConn{
		readWriter: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		conn:       conn,
		connType:   connType,
		logger:     logger,
	}
}

func (c NetworkConn) WriteString(data string) (int, error) {
	numBytes, err := c.readWriter.Write([]byte(data))
	if err != nil {
		return 0, nil
	}

	if err := c.readWriter.Flush(); err != nil {
		return 0, nil
	}
	return numBytes, nil
}

func getBulkStringLength(rawStr string) (int64, error) {
	bulkStringSizeWithPrefix := strings.TrimSuffix(rawStr, "\r\n")
	bulkStringSize, err := command.ParseIntWithPrefix(bulkStringSizeWithPrefix, "$")
	if err != nil {
		return 0, fmt.Errorf("Error parsing bulk strings size %q to int: %w", bulkStringSizeWithPrefix, err)
	}
	return bulkStringSize, nil

}

func (c NetworkConn) ReadNextCmdString() (string, error) {
	firstReadRes, err := c.readWriter.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(firstReadRes) == 0 {
		return "", nil
	}
	c.logger.Info("reading in first part of command string", zap.String("result", firstReadRes))

	switch (firstReadRes)[0] {
	case '*':
		var sb strings.Builder

		// If we have an array string, we need to continue reading the rest of the elements
		arraySizeStrWithPrefix := strings.TrimSuffix(firstReadRes, "\r\n")
		arraySize, err := command.ParseIntWithPrefix(arraySizeStrWithPrefix, "*")
		if err != nil {
			return "", fmt.Errorf("Error parsing array size %q to int: %w", arraySizeStrWithPrefix, err)
		}

		sb.WriteString(firstReadRes)
		for idx := range arraySize { // +1 to include the terminating \r\n
			nextRes, err := c.ReadNextCmdString()
			if err != nil {
				return "", fmt.Errorf("Error reading array entry %d from a bulk string with length %d: %w", arraySize, idx, err)
			}
			sb.WriteString(nextRes)

		}
		return sb.String(), nil
	case '$':
		bulkStringSize, err := getBulkStringLength(firstReadRes)
		if err != nil {
			return "", fmt.Errorf("failed to parse bulk string size: %w", err)
		}

		bulkString, err := c.readWriter.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading %d bytes of bulk string: %w", bulkStringSize, err)
		}
		if int(bulkStringSize)+len("\r\n") != len(bulkString) {
			return "", fmt.Errorf("tried to read bulk string of size %d, but got %d bytes", bulkStringSize, len(bulkString))
		}

		return firstReadRes + bulkString, nil
	default:
	}

	return firstReadRes, nil
}

func (c NetworkConn) ReadRDBFile() (string, error) {
	firstReadRes, err := c.readWriter.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(firstReadRes) == 0 {
		return "", errors.New("received empty RDB file")
	}

	switch (firstReadRes)[0] {
	case '$':
		bulkStringSize, err := getBulkStringLength(firstReadRes)
		if err != nil {
			return "", fmt.Errorf("failed to parse bulk string size: %w", err)
		}

		// The RDB file will not be terminated with a \r\n
		bulkStringBytes := make([]byte, bulkStringSize)
		numBytesRead, err := c.readWriter.Read(bulkStringBytes)
		if err != nil {
			return "", fmt.Errorf("error reading %d bytes of bulk string: %w", bulkStringSize, err)
		}
		if bulkStringSize != int64(numBytesRead) {
			return "", fmt.Errorf("tried to read bulk string of size %d, but only got %d bytes", bulkStringSize, numBytesRead)
		}

		return firstReadRes + string(bulkStringBytes), nil
	default:
	}

	return "", fmt.Errorf("tried to read off RDB file, but got %q", firstReadRes)
}

func (c NetworkConn) ConnectionType() ConnectionType {
	return c.connType
}

func (c NetworkConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c NetworkConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c NetworkConn) Close() error {
	return c.conn.Close()
}
