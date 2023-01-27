package packets

import (
	"bufio"
	"io"
)

func ReadPacketFromConnection(connectionReader *bufio.Reader) ([]byte, error) {
	packetTypeAndFlags, err := connectionReader.Peek(1)

	if err != nil && len(packetTypeAndFlags) == 0 {
		return nil, err
	}

	var header []byte

	for i := 0; i < 4; i++ {
		h, _ := connectionReader.Peek(2 + i)
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
