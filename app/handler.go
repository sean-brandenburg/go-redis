package main

import (
	"context"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

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
			fmt.Println("client handler exiting: ", ctx.Err())
			return
		default:
		}

		data := make([]byte, 128)
		bytesRead, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from client connection: ", err.Error())
			return
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
			fmt.Println("event loop exiting: ", ctx.Err())
			return
		case event := <-s.events:
			cmd, err := command.ParseCommand(event.Command)
			if err != nil {
				fmt.Println("Error parsing client command: ", err.Error())
				continue
			}

			fmt.Println("Executing command: ", cmd.String())

			err = cmd.ExecuteCommand(event.ClientConn)
			if err != nil {
				fmt.Println("Error executing client command", err.Error())
				continue
			}
		}
	}
}
