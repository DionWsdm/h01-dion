package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestCodecEncodingDecoding(t *testing.T) {
	messageType := 0x1    // 4 bits
	messagePurpose := 0x5 // 4 bits
	messageBody := "Test Payload"

	headers := uint8((messageType << 4) | messagePurpose)

	expectedFixed := MessageFixedSegment{
		ID:            42,
		Headers:       headers,
		MessageLength: uint8(len(messageBody)),
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, expectedFixed); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	buf.WriteString(messageBody)

	encodedBytes := buf.Bytes()
	if len(encodedBytes) != 4+len(messageBody) {
		t.Fatalf("unexpected encoded size: got %d, want %d", len(encodedBytes), 4+len(messageBody))
	}

	reader := bytes.NewReader(encodedBytes)
	var decodedFixed MessageFixedSegment
	if err := binary.Read(reader, binary.BigEndian, &decodedFixed); err != nil {
		t.Fatalf("failed to decode header: %v", err)
	}

	if decodedFixed.ID != expectedFixed.ID {
		t.Errorf("ID mismatch: got %d, want %d", decodedFixed.ID, expectedFixed.ID)
	}
	if decodedFixed.Headers != expectedFixed.Headers {
		t.Errorf("Headers mismatch: got %08b, want %08b", decodedFixed.Headers, expectedFixed.Headers)
	}

	decodedBody := make([]byte, decodedFixed.MessageLength)
	if _, err := io.ReadFull(reader, decodedBody); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if string(decodedBody) != messageBody {
		t.Errorf("Body mismatch: got %q, want %q", string(decodedBody), messageBody)
	}

	decodedPurpose := int(decodedFixed.Headers) & 0x0F
	decodedType := int(decodedFixed.Headers) >> 4

	if decodedType != messageType {
		t.Errorf("Type mismatch: got %d, want %d", decodedType, messageType)
	}
	if decodedPurpose != messagePurpose {
		t.Errorf("Purpose mismatch: got %d, want %d", decodedPurpose, messagePurpose)
	}
}
