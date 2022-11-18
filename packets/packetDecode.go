package packets

import (
	"errors"
	"strings"
)

var (
	errPacketLength error = errors.New("cannot decode packet: packet too short to be a connect packet")
	errInvalidType  error = errors.New("cannot decode packet: invalid control type")
)

// DecodeFixedHeader takes a list of bytes and decodes them to a fixed
// control header. It returns a pointer to a ControlHeader, the length
// of the fixed header in btyes and a potential error.
func DecodeFixedHeader(toDecode []byte) (*ControlHeader, int, error) {
	resultHeader := &ControlHeader{}

	if len(toDecode) < 2 {
		return &ControlHeader{}, 0, errPacketLength
	}

	resultHeader.Type = (toDecode[0] >> 4)

	if !(isValidControlType(resultHeader.Type)) {
		return &ControlHeader{}, 0, errInvalidType
	}

	// Mask out parts of the flag field and get the contents
	// We need to shift QoS to be the first 2 bits
	resultHeader.Retain = (toDecode[0] & 1) == 1
	resultHeader.Dup = (toDecode[0] & 8) == 1
	resultHeader.Qos = (toDecode[0] & 6) >> 1

	fixedLength, varLengthLen, err := DecodeVarLengthInt(toDecode[1:])
	resultHeader.FixedLength = fixedLength

	if err != nil {
		return &ControlHeader{}, 0, err
	}

	return resultHeader, 1 + varLengthLen, nil
}

var (
	errMalformedUTFString = errors.New("malformed UTF string")
)

// FetchUTFString fetches a UTF string as encoded by the MQTT
// standard. First we get the string length from the first 2 bytes
// then we fetch the string itself.
// Returns the decoded string, the total length of this section
// including the two bytes encoding the length and a potential error.
func FetchUTFString(toFetch []byte) (string, int, error) {
	var stringBuilder strings.Builder

	var stringLen int = int(toFetch[1]) + int(toFetch[0])<<8
	if !(0 <= stringLen && stringLen <= 65535) {
		return "", 0, errMalformedUTFString
	}

	_, err := stringBuilder.Write(toFetch[2 : 2+stringLen])
	if err != nil {
		return "", 0, err
	}

	return stringBuilder.String(), 2 + stringLen, nil
}

// type ConnectVariableLengthHeader struct {
// 	ProtocolName  string
// 	ProtocolLevel byte
// 	ConnectFlags  byte
// 	KeepAlive     int
// }

// Should give us back a packet or throw an error
// We'll pass a slice of the packet
func DecodeConnect(packet []byte) (*Packet, error) {

	resultPacket := &Packet{}

	// Handle the fixed length header
	fixedHeader, fixedHeaderLen, err := DecodeFixedHeader(packet)
	if err != nil {
		return &Packet{}, err
	}
	resultPacket.ControlHeader = fixedHeader

	// Handle the variable length header
	varHeaderDecode := packet[fixedHeaderLen:]
	varHeader := &ConnectVariableLengthHeader{}

	protocolName, offset, err := FetchUTFString(varHeaderDecode)
	if err != nil {
		return &Packet{}, err
	}
	varHeader.ProtocolName = protocolName

	protocolLevel, offset := varHeaderDecode[offset], offset+1
	varHeader.ProtocolLevel = protocolLevel
	flags, offset := varHeaderDecode[offset], offset+1

	varHeader.UsernameFlag = (flags>>7)&1 == 1
	varHeader.PasswordFlag = (flags>>6)&1 == 1
	varHeader.WillRetainFlag = (flags>>5)&1 == 1
	varHeader.WillQoS = (flags >> 3) & 3
	varHeader.WillFlag = (flags>>2)&1 == 1
	varHeader.CleanSession = (flags>>1)&1 == 1

	keepAlive := int(varHeaderDecode[offset])<<8 + int(varHeaderDecode[offset+1])
	varHeader.KeepAlive = keepAlive
	offset += 2

	// If all goes well, we can return
	return resultPacket, nil
}

var errMalformedInt error = errors.New("malformed variable length integer")

// DecodeVarLengthInt takes a list of bytes and decodes a variable length
// header contained in the first 4 bytes. This works according to the
// MQTT Spec for fixed length headers.
// Returns the encoded int, the length of the fixed length header in bytes
// and a potential error.
func DecodeVarLengthInt(packet []byte) (int, int, error) {
	var result int = 0
	multiplier := 1
	length := 0

	for i, val := range packet {
		if i > 3 {
			return 0, 0, errMalformedInt
		}
		// We don't want to include the bit that just signals
		// if we should continue decoding
		result += (int(val) - 128*int(128&val)) * multiplier
		multiplier *= 128

		// We continue decoding if the first bit is a 1, otherwise we stop
		if 128&val == 0 {
			length = i + 1
			break
		}
	}
	return result, length, nil
}
