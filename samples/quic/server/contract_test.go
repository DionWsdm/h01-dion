package main

import (
	"net"
	"testing"

	"github.com/quic-go/quic-go"
)

// Compile-time contract assertions: verifies that all initial variables,
// functions, and signatures remain intact. Students may add additional
// helper functions or variables, provided existing ones are preserved.
var (
	_ string                       = DefaultServerIP
	_ string                       = DefaultServerPort
	_ string                       = ServerType
	_ int                          = BufferSize
	_ string                       = AppLayerProto
	_ string                       = LogDir
	_ string                       = SSLKeyLogFileName
	_ func() (string, string)      = ResolveConfig
	_ func() string                = ResolveALPN
	_ func()                       = main
	_ func(*quic.Conn)             = connectionHandler
	_ func(net.Addr, *quic.Stream) = streamHandler
	_ func(string) (string, error) = logic
)

func TestContract(t *testing.T) {
	// Satisfied at compile time.
}
