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
