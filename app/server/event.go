package server

import "net"

type Event struct {
	// The event string to be handled
	Command string

	// The client connection that the server should send the response to
	ClientConn net.Conn

	// If true, the server should send a response to the client. Otherwise, the server should
	// execute the command without sending anything back to the caller
	ShouldRespond bool
}
