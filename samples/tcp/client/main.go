package main

import (
	"bufio"
	"compnet-socket-labs/samples/r1"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	DefaultServerIP   = r1.DefaultServerPublicIp
	DefaultServerPort = r1.DefaultServerPort
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
	fmt.Printf("[%s] Connecting to %s\n", ServerType, net.JoinHostPort(serverIP, serverPort))

	remoteTCPAddress, err := net.ResolveTCPAddr(ServerType, net.JoinHostPort(serverIP, serverPort))
	if err != nil {
		log.Fatalln(err)
	}
	socket, err := net.DialTCP(ServerType, nil, remoteTCPAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer socket.Close()

	fmt.Printf("TCP Client Socket Program Example in Go\n")
	fmt.Printf("[%s] Dialling from %s to %s\n", ServerType, socket.LocalAddr(), socket.RemoteAddr())

	fmt.Printf("[%s] Creating receive buffer of size %d\n", ServerType, BufferSize)
	receiveBuffer := make([]byte, BufferSize)

	fmt.Printf("[%s] Input message to be sent to server: ", ServerType)
	message, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("[%s] Sending message '%s' to server\n", ServerType, message)
	_, err = socket.Write([]byte(message))
	if err != nil {
		log.Fatalln(err)
	}

	receiveLength, err := socket.Read(receiveBuffer)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	fmt.Printf("[%s] Received %d bytes of message from server\n", ServerType, receiveLength)

	response := string(receiveBuffer[:receiveLength])
	fmt.Printf("[%s] Response from server: %s\n", ServerType, response)
}
