package main

import "testing"

// Compile-time contract assertions: verifies that all initial variables,
// functions, and signatures remain intact. Students may add additional
// helper functions or variables, provided existing ones are preserved.
var (
	_ string                  = DefaultServerIP
	_ string                  = DefaultServerPort
	_ int                     = BufferSize
	_ string                  = AppLayerProto
	_ func() (string, string) = ResolveConfig
	_ func() string           = ResolveALPN
	_ func()                  = main
)

func TestContract(t *testing.T) {
	// Satisfied at compile time.
}
