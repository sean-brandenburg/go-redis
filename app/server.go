package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"net"
	"os"
)

type Server struct {
	events   chan Event
	listener net.Listener
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379", err.Error())
		os.Exit(1)
	}
	server := Server{
		events:   make(chan Event, 0),
		listener: listener,
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	go server.eventLoop(ctx)
	go server.connectionHandler(ctx)
	
	sigShutdown := make(chan os.Signal, 1)
	signal.Notify(sigShutdown, syscall.SIGTERM, syscall.SIGINT)
	
	select {
		case <-sigShutdown:
			cancel()	
			fmt.Println("server received shutdown signal")
	}
}

func (s Server) connectionHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Errorf("connection handler exiting: ", ctx.Err())
			return
		default:
		}

		clientConn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		go s.clientHandler(ctx, clientConn)
	}
}

func (s Server) clientHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Errorf("client handler exiting: ", ctx.Err())
			return
		default:
		}

		data := make([]byte, 128)
		bytesRead, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from client connection: ", err.Error())
		}

		command := string(data[:bytesRead])
		fmt.Println("Received message: ", command)

		s.events <- Event{
			Command:    command,
			ClientConn: conn,
		}
	}
}

func (s Server) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Errorf("event loop exiting: ", ctx.Err())
			return
		case event := <-s.events:
			fmt.Println("Handling Command: ", event.Command)
			
			respString := "+PONG\r\n"
			fmt.Println("Writing message: ", respString)
			
			_, err := event.ClientConn.Write([]byte(respString))
			if err != nil {
				fmt.Println("Error writing response to client", err.Error())
				continue
			}
		}
	}
}
