package packets

import (
	"MQTT-GO/structures"
	"bufio"
	"io"
	"sync/atomic"
)

var (
	called = atomic.Int32{}
)

func ReadPacketFromConnection(connectionReader *bufio.Reader) ([]byte, error) {

	structures.Println("Reached here:", called.Add(1))
	packetTypeAndFlags, err := connectionReader.Peek(1)

	if err != nil && len(packetTypeAndFlags) == 0 {
		return nil, err
	}

	var header []byte

	for i := 0; i < 4; i++ {
		h, err := connectionReader.Peek(2 + i)
		if err != nil {
			return nil, err
		}
		// Read until we've reached the end of the var length int
		if h[i+1]&128 == 0 {
			header = h
			break
		}
	}

	dataLen, varLengthIntLen, err := DecodeVarLengthInt(header[1:])
	if err != nil {
		return nil, err
	}

	packet := make([]byte, dataLen+varLengthIntLen+1)
	bytesRead, err := io.ReadFull(connectionReader, packet)
	packet = packet[:bytesRead]
	if err != nil {
		return nil, err
	}

	return packet, nil
}
