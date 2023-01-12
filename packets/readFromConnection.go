package packets

import (
	"bufio"
	"io"
)

func ReadPacketFromConnection(connectionReader *bufio.Reader) ([]byte, error) {
	buffer, err := connectionReader.Peek(4)
	if err != nil && len(buffer) == 0 {
		return nil, err
	}

	dataLen, varLengthIntLen, err := DecodeVarLengthInt(buffer[1:])
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
