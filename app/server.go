package main

import (
	"fmt"

	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379", err.Error())
		os.Exit(1)
	}

	clientConn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer clientConn.Close()

	data := make([]byte, 128)
	bytesRead, err := clientConn.Read(data)
	if err != nil {
		fmt.Println("Error reading from client connection: ", err.Error())
		os.Exit(1)
	}
	
	fmt.Printf("Read %d bytes: \n", bytesRead)
	fmt.Println("Received message: ", string(data[:bytesRead]))

	respString := "+PONG\r\n"
	// for range 2 {
		fmt.Println("Writing message: ", respString)
		_, err = clientConn.Write([]byte(respString))
		if err != nil {
			fmt.Println("Error writing response to client", err.Error())
			os.Exit(1)
		}
	// }
}
