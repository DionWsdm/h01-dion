package main

import (
	"compnet-socket-labs/samples/r1"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	DefaultServerIP   = r1.DefaultServerPrivateIp
	DefaultServerPort = r1.DefaultServerPort
	ServerType        = "udp4"
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
	listenAddress, err := net.ResolveUDPAddr(ServerType, net.JoinHostPort(serverIP, serverPort))
	if err != nil {
		log.Fatalln(err)
	}

	socket, err := net.ListenUDP(ServerType, listenAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer socket.Close()

	fmt.Printf("UDP Server Socket Program Example in Go\n")
	fmt.Printf("Press Ctrl+C or Cmd+C to stop the program\n")
	fmt.Printf("[%s] Listening on: %s\n", ServerType, socket.LocalAddr())

	for {
		fmt.Printf("[%s] Creating receive buffer for next communication of size %d\n", ServerType, BufferSize)
		receiveBuffer := make([]byte, BufferSize)

		receiveLength, address, err := socket.ReadFromUDP(receiveBuffer)
		if err != nil {
			log.Printf("[%s] Read error: %v\n", ServerType, err)
			continue
		}

		go connectionHandler(socket, address, receiveBuffer, receiveLength)
	}
}

func connectionHandler(socket *net.UDPConn, address *net.UDPAddr, receiveBuffer []byte, receiveLength int) {
	fmt.Printf("[%s] [Client: %s] Received %d bytes of message\n", ServerType, address, receiveLength)
	message := string(receiveBuffer[:receiveLength])

	fmt.Printf("[%s] [Client: %s] Message: %s\n", ServerType, address, message)

	response, err := logic(message)
	if err != nil {
		log.Printf("[%s] [Client: %s] Logic error: %v\n", ServerType, address, err)
		return
	}

	fmt.Printf("[%s] [Client: %s] Sending Response: %s\n", ServerType, address, response)
	_, err = socket.WriteToUDP([]byte(response), address)
	if err != nil {
		log.Printf("[%s] [Client: %s] Write error: %v\n", ServerType, address, err)
		return
	}
}

func logic(input string) (string, error) {
	return strings.ToUpper(input), nil
}
