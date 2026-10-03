package utils

import (
	"math"
)

var ()

type LRTJPIDSPacketFixed struct {
	TransactionId          uint16
	IsAck                  bool
	IsNewTrain             bool
	IsUpdateTrain          bool
	IsDeleteTrain          bool
	IsTrainArriving        bool
	IsTrainDeparting       bool
	TrainNumber            uint16
	EstimatedArrivalHour   uint8
	EstimatedArrivalMinute uint8
	DestinationLength      uint8
}

type LRTJPIDSPacket struct {
	LRTJPIDSPacketFixed
	Destination string
}

func getBytes(num uint16) (byte, byte) {
	firstByte := math.Floor(float64(num / 256))
	secondByte := num % 256

	return byte(firstByte), byte(secondByte)
}

func getUint16(b1 byte, b2 byte) uint16 {
	return uint16(b1)*256 + uint16(b2)
}

func getFlagByte(
	IsAck bool,
	IsNewTrain bool,
	IsUpdateTrain bool,
	IsDeleteTrain bool,
	IsTrainArriving bool,
	IsTrainDeparting bool) byte {

	result := 0
	if IsAck {
		result += 1 << 7
	}
	if IsNewTrain {
		result += 1 << 6
	}
	if IsUpdateTrain {
		result += 1 << 5
	}
	if IsDeleteTrain {
		result += 1 << 4
	}
	if IsTrainArriving {
		result += 1 << 3
	}
	if IsTrainDeparting {
		result += 1 << 2
	}
	return byte(result)
}

func getFlagBools(flagByte byte) (bool, bool, bool, bool, bool, bool) {
	IsAck := false
	IsNewTrain := false
	IsUpdateTrain := false
	IsDeleteTrain := false
	IsTrainArriving := false
	IsTrainDeparting := false
	if flagByte >= 128 {
		IsAck = true
		flagByte -= 128
	}
	if flagByte >= 64 {
		IsNewTrain = true
		flagByte -= 64
	}
	if flagByte >= 32 {
		IsUpdateTrain = true
		flagByte -= 32
	}
	if flagByte >= 16 {
		IsDeleteTrain = true
		flagByte -= 16
	}
	if flagByte >= 8 {
		IsTrainArriving = true
		flagByte -= 8
	}
	if flagByte >= 4 {
		IsTrainDeparting = true
		flagByte -= 4
	}
	return IsAck, IsNewTrain, IsUpdateTrain, IsDeleteTrain, IsTrainArriving, IsTrainDeparting
}

func Encoder(packet LRTJPIDSPacket) []byte {
	size := 8 + packet.LRTJPIDSPacketFixed.DestinationLength
	encoded := make([]byte, size)

	encoded[0], encoded[1] = getBytes(packet.TransactionId)
	encoded[2] = getFlagByte(packet.IsAck, packet.IsNewTrain, packet.IsUpdateTrain, packet.IsDeleteTrain, packet.IsTrainArriving, packet.IsTrainDeparting)
	encoded[3], encoded[4] = getBytes(packet.TrainNumber)
	encoded[5] = packet.EstimatedArrivalHour
	encoded[6] = packet.EstimatedArrivalMinute
	encoded[7] = packet.DestinationLength

	bytedDestination := []byte(packet.Destination)
	for i := 0; i < int(packet.DestinationLength); i++ {
		encoded[8+i] = bytedDestination[i]
	}

	return encoded
}

func Decoder(rawMessage []byte) LRTJPIDSPacket {
	IsAck, IsNewTrain, IsUpdateTrain, IsDeleteTrain, IsTrainArriving, IsTrainDeparting := getFlagBools(rawMessage[2])

	packetFixed := LRTJPIDSPacketFixed{
		TransactionId:          getUint16(rawMessage[0], rawMessage[1]),
		IsAck:                  IsAck,
		IsNewTrain:             IsNewTrain,
		IsUpdateTrain:          IsUpdateTrain,
		IsDeleteTrain:          IsDeleteTrain,
		IsTrainArriving:        IsTrainArriving,
		IsTrainDeparting:       IsTrainDeparting,
		TrainNumber:            getUint16(rawMessage[3], rawMessage[4]),
		EstimatedArrivalHour:   rawMessage[5],
		EstimatedArrivalMinute: rawMessage[6],
		DestinationLength:      rawMessage[7],
	}
	packet := LRTJPIDSPacket{
		LRTJPIDSPacketFixed: packetFixed,
		Destination:         string(rawMessage[8:]),
	}
	return packet
}
