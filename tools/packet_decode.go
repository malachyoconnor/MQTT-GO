package tools

import (
	"errors"
	"strings"
)

var (
	errPacketLength error = errors.New("cannot decode packet: packet too short to be a connect packet")
	errInvalidType  error = errors.New("cannot decode packet: invalid control type")
)

// DecodeFixedHeader takes a list of bytes and decodes them to a packet.
// This can be used to decode multiple different packet types as
// this fixed header is common across all of them.
func DecodeFixedHeader(toDecode []byte) (*ControlHeader, error) {
	resultHeader := &ControlHeader{}

	if len(toDecode) < 2 {
		return &ControlHeader{}, errPacketLength
	}

	resultHeader.Type = (toDecode[0] >> 4)

	if !(isValidControlType(resultHeader.Type)) {
		return &ControlHeader{}, errInvalidType
	}

	// Mask out parts of the flag field and get the contents
	// We need to shift QoS to be the first 2 bits
	resultHeader.Retain = (toDecode[0] & 1) == 1
	resultHeader.Dup = (toDecode[0] & 8) == 1
	resultHeader.Qos = (toDecode[0] & 6) >> 1

	fixedLength, err := DecodeVarLengthInt(toDecode[1:])
	resultHeader.FixedLength = fixedLength

	if err != nil {
		return &ControlHeader{}, err
	}

	return resultHeader, nil
}

var (
	errMalformedUTFString = errors.New("malformed UTF string")
)

// FetchUTFString fetches a UTF string as encoded by the MQTT
// standard. First we get the string length from the first 2 bytes
// then we fetch the string itself.
func FetchUTFString(toFetch []byte) (string, error) {
	var stringBuilder strings.Builder

	var stringLen int = int(toFetch[1]) + int(toFetch[0])<<8
	if !(0 <= stringLen && stringLen <= 65535) {
		return "", errMalformedUTFString
	}

	_, err := stringBuilder.Write(toFetch[2 : 2+stringLen])
	if err != nil {
		return "", err
	}

	return stringBuilder.String(), nil
}

// Should give us back a packet or throw an error
// We'll pass a slice of the packet
func DecodeConnect(toDecode []byte) (*Packet, error) {

	resultPacket := &Packet{}

	// Handle the fixed length header
	fixedHeader, err := DecodeFixedHeader(toDecode)
	if err != nil {
		return &Packet{}, err
	}

	resultPacket.ControlHeader = fixedHeader

	// Handle the variable length header

	// If all goes well, we can return
	return resultPacket, nil
}

var errMalformedInt error = errors.New("malformed variable length integer")

func DecodeVarLengthInt(packet []byte) (int, error) {
	var result int = 0
	multiplier := 1

	for i, val := range packet {
		if i > 3 {
			return 0, errMalformedInt
		}
		// We don't want to include the bit that just signals
		// if we should continue decoding
		result += (int(val) - 128*int(128&val)) * multiplier
		multiplier *= 128

		// We continue decoding if the first bit is a 1, otherwise we stop
		if 128&val == 0 {
			break
		}
	}

	return result, nil
}
