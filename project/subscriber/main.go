package main

import (
	"os"

	"compnet-socket-labs/project/utils"
)

var (
	DefaultServerIP   = "127.0.0.1"
	DefaultServerPort = "54321"
	ServerType        = "udp4"
	BufferSize        = 2048
	AppLayerProto     = "lrt-jakarta-demo"
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

func Handler(packet utils.LRTJPIDSPacket) string {
	return ""
}

func main() {
	_ = Handler(utils.LRTJPIDSPacket{})
}
