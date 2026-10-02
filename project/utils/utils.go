package utils

// Component data types may be modified, but field names must not be changed
type LRTJPIDSPacketFixed struct {
	TransactionId          int
	IsAck                  int
	IsNewTrain             int
	IsUpdateTrain          int
	IsDeleteTrain          int
	IsTrainArriving        int
	IsTrainDeparting       int
	TrainNumber            int
	EstimatedArrivalHour   int
	EstimatedArrivalMinute int
	DestinationLength      int
}

type LRTJPIDSPacket struct {
	LRTJPIDSPacketFixed
	Destination string
}

func Encoder(packet LRTJPIDSPacket) []byte {
	return nil
}

func Decoder(rawMessage []byte) LRTJPIDSPacket {
	return LRTJPIDSPacket{}
}
