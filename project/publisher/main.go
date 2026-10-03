package main

import (
	"compnet-socket-labs/project/config"
	"compnet-socket-labs/project/utils"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/qlog"
)

var (
	DefaultServerIP   = config.DefaultSubscriberPublicIP
	DefaultServerPort = config.DefaultServerPort
	BufferSize        = config.BufferSize
	AppLayerProto     = config.AppLayerProto
	ServerType        = config.ServerType
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

func SendPacket(connection *quic.Conn, packet utils.LRTJPIDSPacket) {
	packet.IsAck = false // make sure IsACK false karena ini pengiriman dari client
	encodedPacket := utils.Encoder(packet)

	stream, err := connection.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	defer stream.Close()

	fmt.Printf("[quic] Opened bidirectional stream to %s\n", connection.RemoteAddr())

	_, err = stream.Write([]byte(encodedPacket))
	if err != nil {
		log.Fatalln(err)
	}

	receiveBuffer := make([]byte, BufferSize)

	_, err = stream.Read(receiveBuffer)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	decodedResponse := utils.Decoder(receiveBuffer)
	fmt.Printf("[quic] Decoded response from server: %v\n", decodedResponse)
}

func main() {
	destinationA := "Manggarai"
	packetAFixed := utils.LRTJPIDSPacketFixed{
		TransactionId:          1,
		IsAck:                  false,
		IsNewTrain:             true,
		IsUpdateTrain:          false,
		IsDeleteTrain:          false,
		IsTrainArriving:        false,
		IsTrainDeparting:       false,
		TrainNumber:            1002,
		DestinationLength:      uint8(len(destinationA)),
		EstimatedArrivalHour:   5,
		EstimatedArrivalMinute: 34,
	}
	packetA := utils.LRTJPIDSPacket{
		LRTJPIDSPacketFixed: packetAFixed,
		Destination:         destinationA,
	}

	destinationB := "Velodrome"
	packetBFixed := utils.LRTJPIDSPacketFixed{
		TransactionId:          2,
		IsAck:                  false,
		IsNewTrain:             false,
		IsUpdateTrain:          true,
		IsDeleteTrain:          false,
		IsTrainArriving:        false,
		IsTrainDeparting:       false,
		TrainNumber:            1002,
		EstimatedArrivalHour:   5,
		EstimatedArrivalMinute: 35,
		DestinationLength:      uint8(len(destinationB)),
	}
	packetB := utils.LRTJPIDSPacket{
		LRTJPIDSPacketFixed: packetBFixed,
		Destination:         destinationB,
	}

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

	SendPacket(connection, packetA)
	SendPacket(connection, packetB)
}
