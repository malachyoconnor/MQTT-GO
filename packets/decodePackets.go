package packets

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	errPacketTooShort error = errors.New("error: cannot decode packet: packet too short to be a connect packet")
	errInvalidType    error = errors.New("error: cannot decode packet: invalid control type")
	errInvalidLength  error = errors.New("error: packet length differs from the advertised fixed length")
)

func PrintPacket(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

// DecodeFixedHeader takes a packet and decodes the fixed header.
// It returns a pointer to a ControlHeader, the length
// of the fixed header in btyes and a potential error.
func DecodeFixedHeader(packet []byte) (*ControlHeader, int, error) {
	resultHeader := &ControlHeader{}

	if len(packet) < 2 {
		return &ControlHeader{}, 0, errPacketTooShort
	}

	resultHeader.Type = (packet[0] >> 4)

	if !(isValidControlType(resultHeader.Type)) {
		return &ControlHeader{}, 0, errInvalidType
	}

	// Mask out parts of the flag field and get the contents
	// We need to shift QoS to be the first 2 bits
	resultHeader.Retain = (packet[0] & 1) == 1
	resultHeader.Dup = (packet[0] & 8) == 1
	resultHeader.Qos = (packet[0] & 6) >> 1

	fixedLength, varLengthLen, err := DecodeVarLengthInt(packet[1:])
	resultHeader.RemainingLength = fixedLength
	if err != nil {
		return &ControlHeader{}, 0, err
	}

	if fixedLength != len(packet)-(1+varLengthLen) {
		return &ControlHeader{}, 0, errInvalidLength
	}

	return resultHeader, 1 + varLengthLen, nil
}

var errMalformedUTFString = errors.New("error: malformed UTF string")

// FetchUTFString fetches a UTF string as encoded by the MQTT
// standard. First we get the string length from the first 2 bytes
// then we fetch the string itself.
// Returns the decoded string, the total length of this section
// including the two bytes encoding the length, and a potential error.
func FetchUTFString(toFetch []byte) (string, int, error) {
	var stringBuilder strings.Builder

	var stringLen int = CombineMsbLsb(toFetch[0], toFetch[1])
	if !(0 <= stringLen && stringLen <= 65535) || (stringLen > len(toFetch)-2) {
		return "", 0, errMalformedUTFString
	}

	_, err := stringBuilder.Write(toFetch[2 : 2+stringLen])
	if err != nil {
		return "", 0, err
	}

	return stringBuilder.String(), 2 + stringLen, nil
}

var (
	errShrunkenByteArr error = errors.New("error: input byte string to FetchBytes was too short")
)

// FetchBytes fetches as many bytes as given by the first two bytes
// in an input byte array (excluding the first 2 bits (the length itself)).
// Returns the fetched bytes, the total length of this section
// including the two bytes encoding the length, and a potential error.
func FetchBytes(toFetch []byte) ([]byte, int, error) {
	var numBytes int = CombineMsbLsb(toFetch[0], toFetch[1])
	if len(toFetch) < numBytes+2 {
		return []byte{}, 0, errShrunkenByteArr
	}
	resultArr := make([]byte, numBytes)
	copy(resultArr, toFetch[2:2+numBytes])

	return resultArr, 2 + numBytes, nil
}

var (
	errPacketNotDefined error = errors.New("error: Packet type not defined")
)

func GetPacketType(packet *[]byte) byte {
	return (*packet)[0] >> 4
}

var zeroLengthPacketError = errors.New("error: Zero length packet read from byte pool.")

// DecodePacket takes a byte array encoding a packet and returns
// (*Packet, PacketType, error)
func DecodePacket(packet []byte) (*Packet, byte, error) {

	if len(packet) == 0 {
		return &Packet{}, 0, zeroLengthPacketError
	}

	packetType := GetPacketType(&packet)

	var result *Packet
	var err error

	switch packetType {

	case CONNECT:
		result, err = DecodeConnect(packet[:])

	case PUBLISH:
		result, err = DecodePublish(packet[:])

		var stringBuilder strings.Builder
		stringBuilder.Write(result.Payload.ApplicationMessage)
		fmt.Println("Decoded:", stringBuilder.String())

	case PINGREQ:
		result, err = DecodePing(packet[:])
		fmt.Println("Ping")

	default:

		fmt.Println("Packet type not defined: ", packetType, " (", packetTypeName(packetType), ")")
		return &Packet{}, 0, errPacketNotDefined
	}

	if err != nil {
		return &Packet{}, 0, err
	}
	return result, packetType, nil
}

// Should give us back a packet or throw an error
// We'll pass a slice of the packet
func DecodeConnect(packet []byte) (*Packet, error) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("Recovered from", r)
		}
	}()
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, fixedHeaderLen, err := DecodeFixedHeader(packet)

	if err != nil {
		return &Packet{}, err
	}

	if fixedHeader.Type != 1 {
		err := fmt.Errorf("error: Incorrect packet type. Given type %v to connect ", fixedHeader.Type)
		return &Packet{}, err
	}

	resultPacket.ControlHeader = *fixedHeader

	// Handle the variable length header
	varHeaderDecode := packet[fixedHeaderLen:]
	varHeader := ConnectVariableLengthHeader{}

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

	keepAlive := CombineMsbLsb(varHeaderDecode[offset], varHeaderDecode[offset+1])
	varHeader.KeepAlive = keepAlive
	offset += 2

	resultPacket.VariableLengthHeader = varHeader

	// Now we've to decode the payload
	resultPayload := PacketPayload{}
	payloadDecode := varHeaderDecode[offset:]

	clientID, offset, err := FetchUTFString(payloadDecode)
	if err != nil {
		return &Packet{}, err
	}
	resultPayload.clientId = clientID

	if varHeader.WillFlag {
		willTopic, addedOffset, err := FetchUTFString(payloadDecode[offset:])
		if err != nil {
			return &Packet{}, err
		}
		offset += addedOffset

		willMessage, addedOffset, err := FetchBytes(payloadDecode[offset:])
		if err != nil {
			return &Packet{}, err
		}
		offset += addedOffset

		resultPayload.willTopic = willTopic
		resultPayload.willMessage = willMessage
	}

	if varHeader.UsernameFlag {
		username, addedOffset, err := FetchUTFString(payloadDecode[offset:])
		if err != nil {
			return &Packet{}, err
		}
		offset += addedOffset
		resultPayload.username = username
	}

	if varHeader.PasswordFlag {
		password, addedOffset, err := FetchUTFString(payloadDecode[offset:])
		if err != nil {
			return &Packet{}, err
		}
		offset += addedOffset
		resultPayload.password = password
	}

	resultPacket.Payload = resultPayload
	// If all goes well, we can return
	return resultPacket, nil
}

// CombineMsbLsb takes two bytes (a most significat big and a least) and
// combines them into an int - the combine values are contained in the
// first 16 bits of the result
func CombineMsbLsb(msb byte, lsb byte) int {
	return int(msb)<<8 + int(lsb)
}

var errMalformedInt error = errors.New("error: malformed variable length integer")

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

func DecodePublish(packet []byte) (*Packet, error) {
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, offset, err := DecodeFixedHeader(packet)
	if err != nil {
		return &Packet{}, err
	}
	resultPacket.ControlHeader = *fixedHeader

	// Handle the variable length header
	varHeader := PublishVariableLengthHeader{}
	topicName, topicLen, err := FetchUTFString(packet[offset:])
	varHeaderLen := topicLen

	if fixedHeader.Qos == 1 || fixedHeader.Qos == 2 {
		packetIdentifier := CombineMsbLsb(packet[offset+topicLen], packet[offset+topicLen+1])
		varHeader.PacketIdentifier = packetIdentifier
		varHeaderLen += 2
	}

	if err != nil {
		return &Packet{}, err
	}

	varHeader.TopicName = topicName
	payloadLength := fixedHeader.RemainingLength - varHeaderLen
	offset = offset + varHeaderLen

	resultPacket.VariableLengthHeader = varHeader

	var payload PacketPayload
	payload.ApplicationMessage = make([]byte, payloadLength)
	copy(payload.ApplicationMessage, packet[offset:offset+payloadLength])
	resultPacket.Payload = payload

	return resultPacket, nil

}

func DecodePing(packet []byte) (*Packet, error) {
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, offset, err := DecodeFixedHeader(packet)
	if err != nil {
		return &Packet{}, err
	}
	resultPacket.ControlHeader = *fixedHeader
	if offset != len(packet) {
		return &Packet{}, errInvalidLength
	}

	return resultPacket, nil

}

func CreateByteInline(input_binary []byte) byte {
	res, _ := CreateByte(input_binary)
	return res
}

// CreateByte takes an array of 0s and 1s and returns the byte representation
// along with a boolean "ok".
func CreateByte(input_binary []byte) (byte, bool) {
	if len(input_binary) > 8 {
		return 0, false
	}
	var result byte = 0
	for i, v := range input_binary {
		if !(v == 0 || v == 1) {
			return 0, false
		}
		result += (v << (7 - i))
	}
	return result, true
}
