package main

import (
	"bufio"
	"compnet-socket-labs/samples/r4"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/qlog"
)

var (
	Wg                sync.WaitGroup
	TotalStream       = r4.TotalStream
	DefaultServerIP   = "127.0.0.1"
	DefaultServerPort = "54321"
	ServerType        = "udp4"
	BufferSize        = 2048
	AppLayerProto     = "jarkom-quic-sample-dion"
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

func handleStream(connection *quic.Conn, message string, receiveBuffer []byte) {
	defer Wg.Done()

	stream2, err := connection.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	streamId := stream2.StreamID()

	defer stream2.Close()

	fmt.Printf("[quic] Opened bidirectional stream %d to %s\n", streamId, connection.RemoteAddr())

	fmt.Printf("[quic] [Stream %d] Sending message '%s' to server\n", streamId, message)
	_, err = stream2.Write([]byte(message))
	if err != nil {
		log.Fatalln(err)
	}

	receiveLength, err := stream2.Read(receiveBuffer)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	fmt.Printf("[quic] [Stream %d] Received %d bytes of message from server\n", streamId, receiveLength)

	response := string(receiveBuffer[:receiveLength])
	fmt.Printf("[quic] [Stream %d] Response from server: %s\n", streamId, response)
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

	fmt.Printf("QUIC Client Socket Program Example in Go\n")
	fmt.Printf("[%s] Connecting to %s\n", ServerType, net.JoinHostPort(serverIP, serverPort))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Self-signed test certificate
		NextProtos:         []string{alpn},
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

	connection, err := quic.DialAddr(context.Background(), net.JoinHostPort(serverIP, serverPort), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.CloseWithError(0x0, "No Error")

	fmt.Printf("[quic] Dialling from %s to %s\n", connection.LocalAddr(), connection.RemoteAddr())

	fmt.Printf("[quic] Creating receive buffer of size %d\n", BufferSize)
	receiveBuffer := make([]byte, BufferSize)

	fmt.Printf("[quic] Input message to be sent to server: ")
	message, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}

	for stream := 0; stream < TotalStream; stream++ {
		Wg.Add(1)
		go handleStream(connection, message, receiveBuffer)
	}

	Wg.Wait()
}
