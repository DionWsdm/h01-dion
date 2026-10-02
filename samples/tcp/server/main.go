package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	DefaultServerIP   = "127.0.0.1"
	DefaultServerPort = "54321"
	ServerType        = "tcp4"
	BufferSize        = 2048
)

func ResolveConfig() (string, string) {
	ip := os.Getenv("SERVER_ADDR")
	if ip == "" {
		ip = DefaultServerIP
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultServerPort
	}
	return ip, port
}

func main() {
	serverIP, serverPort := ResolveConfig()
	listenAddress, err := net.ResolveTCPAddr(ServerType, net.JoinHostPort(serverIP, serverPort))
	if err != nil {
		log.Fatalln(err)
	}

	socket, err := net.ListenTCP(ServerType, listenAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer socket.Close()

	fmt.Printf("TCP Server Socket Program Example in Go\n")
	fmt.Printf("Press Ctrl+C or Cmd+C to stop the program\n")
	fmt.Printf("[%s] Listening on: %s\n", ServerType, socket.Addr())

	for {
		connection, err := socket.AcceptTCP()
		if err != nil {
			log.Printf("[%s] Accept error: %v\n", ServerType, err)
			continue
		}

		go connectionHandler(connection)
	}
}

func connectionHandler(connection *net.TCPConn) {
	defer connection.Close()

	fmt.Printf("[%s] Receive connection from %s\n", ServerType, connection.RemoteAddr())
	fmt.Printf("[%s] [Client: %s] Creating receive buffer for connection of size %d\n", ServerType, connection.RemoteAddr(), BufferSize)
	receiveBuffer := make([]byte, BufferSize)

	receiveLength, err := connection.Read(receiveBuffer)
	if err != nil {
		if err != io.EOF {
			log.Printf("[%s] [Client: %s] Read error: %v\n", ServerType, connection.RemoteAddr(), err)
		}
		return
	}

	fmt.Printf("[%s] [Client: %s] Received %d bytes of message\n", ServerType, connection.RemoteAddr(), receiveLength)
	message := string(receiveBuffer[:receiveLength])

	fmt.Printf("[%s] [Client: %s] Message: %s\n", ServerType, connection.RemoteAddr(), message)

	response, err := logic(message)
	if err != nil {
		log.Printf("[%s] [Client: %s] Logic error: %v\n", ServerType, connection.RemoteAddr(), err)
		return
	}

	fmt.Printf("[%s] [Client: %s] Sending Response: %s\n", ServerType, connection.RemoteAddr(), response)
	_, err = connection.Write([]byte(response))
	if err != nil {
		log.Printf("[%s] [Client: %s] Write error: %v\n", ServerType, connection.RemoteAddr(), err)
		return
	}
}

func logic(input string) (string, error) {
	return strings.ToUpper(input), nil
}
