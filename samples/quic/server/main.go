package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/qlog"

	"compnet-socket-labs/samples/quic/quicutil"
)

var (
	DefaultServerIP   = "127.0.0.1"
	DefaultServerPort = "54321"
	ServerType        = "udp4"
	BufferSize        = 2048
	AppLayerProto     = "compnet-quic-sample"
	LogDir            = "logs"
	SSLKeyLogFileName = "ssl-key.log"
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

func ResolveALPN() string {
	alpn := os.Getenv("ALPN")
	if alpn == "" {
		alpn = AppLayerProto
	}
	return alpn
}

func main() {
	keylogFlag := flag.Bool("keylog", false, "Enable TLS key logging to logs/ssl-key.log for Wireshark inspection")
	qlogFlag := flag.Bool("qlog", false, "Enable QLOG event tracing to logs/*.sqlog")
	flag.Parse()

	if *keylogFlag || *qlogFlag {
		if err := os.MkdirAll(LogDir, 0755); err != nil {
			log.Fatalf("failed to create logs directory: %v", err)
		}
	}

	serverIP, serverPort := ResolveConfig()
	alpn := ResolveALPN()

	localUDPAddress, err := net.ResolveUDPAddr(ServerType, net.JoinHostPort(serverIP, serverPort))
	if err != nil {
		log.Fatalln(err)
	}
	socket, err := net.ListenUDP(ServerType, localUDPAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer socket.Close()

	fmt.Printf("QUIC Server Socket Program Example in Go\n")
	fmt.Printf("[%s] Preparing UDP listening socket on %s\n", ServerType, socket.LocalAddr())

	tlsConfig := &tls.Config{
		Certificates: quicutil.GenerateTLSSelfSignedCertificates(),
		NextProtos:   []string{alpn},
	}

	if *keylogFlag {
		keyLogPath := filepath.Join(LogDir, SSLKeyLogFileName)
		keyLogFile, err := os.OpenFile(keyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("failed to open keylog file: %v", err)
		}
		defer keyLogFile.Close()
		tlsConfig.KeyLogWriter = keyLogFile
		fmt.Printf("[quic] TLS Key logging enabled -> %s\n", keyLogPath)
	}

	quicConfig := &quic.Config{}
	if *qlogFlag {
		if os.Getenv("QLOGDIR") == "" {
			_ = os.Setenv("QLOGDIR", LogDir)
		}
		quicConfig.Tracer = qlog.DefaultConnectionTracer
		fmt.Printf("[quic] QLOG event tracing enabled -> %s/*.sqlog\n", LogDir)
	}

	listener, err := quic.Listen(socket, tlsConfig, quicConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	fmt.Printf("[%s] Listening on %s (ALPN: %s)\n", ServerType, socket.LocalAddr(), alpn)
	fmt.Printf("Press Ctrl+C or Cmd+C to stop the program\n")

	for {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("[quic] Accept error: %v\n", err)
			continue
		}

		go connectionHandler(connection)
	}
}

func connectionHandler(connection *quic.Conn) {
	fmt.Printf("[quic] Receive connection from %s\n", connection.RemoteAddr())

	stream, err := connection.AcceptStream(context.Background())
	if err != nil {
		return
	}

	go streamHandler(connection.RemoteAddr(), stream)
}

func streamHandler(remoteAddr net.Addr, stream *quic.Stream) {
	defer stream.Close()

	fmt.Printf("[quic] [Client: %s] Creating receive buffer of size %d\n", remoteAddr, BufferSize)
	receiveBuffer := make([]byte, BufferSize)

	receiveLength, err := stream.Read(receiveBuffer)
	if err != nil {
		if err != io.EOF {
			log.Printf("[quic] [Client: %s] Read error: %v\n", remoteAddr, err)
		}
		return
	}

	fmt.Printf("[quic] [Client: %s] Received %d bytes of message\n", remoteAddr, receiveLength)
	message := string(receiveBuffer[:receiveLength])

	fmt.Printf("[quic] [Client: %s] Message: %s\n", remoteAddr, message)

	response, err := logic(message)
	if err != nil {
		log.Printf("[quic] [Client: %s] Logic error: %v\n", remoteAddr, err)
		return
	}

	fmt.Printf("[quic] [Client: %s] Sending Response: %s\n", remoteAddr, response)
	_, err = stream.Write([]byte(response))
	if err != nil {
		log.Printf("[quic] [Client: %s] Write error: %v\n", remoteAddr, err)
		return
	}
}

func logic(input string) (string, error) {
	return strings.ToUpper(input), nil
}
