package utils

import (
	"crypto/tls"
	"testing"
)

// Compile-time contract assertions: verifies that all initial variables,
// functions, and signatures remain intact. Students may add additional
// helper functions or variables, provided existing ones are preserved.
var (
	_ func(LRTJPIDSPacket) []byte = Encoder
	_ func([]byte) LRTJPIDSPacket = Decoder
	_ func() []tls.Certificate    = GenerateTLSSelfSignedCertificates
)

func TestContract(t *testing.T) {
	// Satisfied at compile time.
}
