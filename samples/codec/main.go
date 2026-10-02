package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type MessageFixedSegment struct {
	ID            uint16
	Headers       uint8
	MessageLength uint8
}

type Message struct {
	MessageFixedSegment
	MessageBody string
}

func main() {
	fmt.Println("=== ENCODING ===")
	messageType := 0x1    // Assume length is 4 bits
	messagePurpose := 0x5 // Assume length is 4 bits

	message := "Hello World"

	fmt.Printf("Message: %s, in bytes-array it is %v\n", message, []byte(message))

	fmt.Printf("%04b %04b\n", messageType, messagePurpose)

	headers := messageType<<4 + messagePurpose

	fmt.Printf("%08b\n", headers)

	sampleMessage := MessageFixedSegment{
		ID:            1234,
		Headers:       uint8(headers),
		MessageLength: uint8(len(message)),
	}

	fmt.Println(sampleMessage)

	bytesBuffer := new(bytes.Buffer)
	if err := binary.Write(bytesBuffer, binary.BigEndian, sampleMessage); err != nil {
		log.Fatalf("failed to encode fixed segment: %v", err)
	}
	bytesBuffer.WriteString(message)

	finalBytes := bytesBuffer.Bytes()
	fmt.Println(finalBytes)

	fmt.Println("=== DECODING ===")

	var decodedFixedSegment MessageFixedSegment
	bytesReader := bytes.NewReader(finalBytes)
	if err := binary.Read(bytesReader, binary.BigEndian, &decodedFixedSegment); err != nil {
		log.Fatalf("failed to decode fixed segment: %v", err)
	}

	decodedMessage := make([]byte, decodedFixedSegment.MessageLength)
	if _, err := io.ReadFull(bytesReader, decodedMessage); err != nil {
		log.Fatalf("failed to read decoded message body: %v", err)
	}

	messagePurpose = int(decodedFixedSegment.Headers) % 16
	messageType = int(decodedFixedSegment.Headers) >> 4

	fmt.Println(decodedFixedSegment)
	fmt.Printf("Decoded message: %s (Type: 0x%X, Purpose: 0x%X)\n", string(decodedMessage), messageType, messagePurpose)
}
