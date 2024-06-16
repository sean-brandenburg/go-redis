package main

import "net"

type Event struct {
	// The event string to be handled
	Command string

	// The client connection that the server should send the response to
	ClientConn net.Conn
}
