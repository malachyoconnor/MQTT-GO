package packets

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errPacketTooShort error = errors.New("error: cannot decode packet: packet too short to be a connect packet")
	errInvalidType    error = errors.New("error: cannot decode packet: invalid control type")
	errInvalidLength  error = errors.New("error: packet length differs from the advertised fixed length")
)

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

	// Mask out the LSF of the header to get the flags
	resultHeader.Flags = packet[0] & 15

	fixedLength, varLengthLen, err := DecodeVarLengthInt(packet[1:])
	resultHeader.RemainingLength = fixedLength
	if err != nil {
		return &ControlHeader{}, 0, err
	}

	if fixedLength != (len(packet)-1)-(varLengthLen) {
		// We still return the values, because we may not have the whole packet yet
		// We may JUST be passing the fixed header
		return resultHeader, 1 + varLengthLen, errors.New("error: packet length differs from the advertised fixed length in DecodeFixedHeader")
	}

	return resultHeader, 1 + varLengthLen, nil
}

var errMalformedUTFString = errors.New("error: malformed UTF string")

// DecodeUTFString fetches a UTF string as encoded by the MQTT
// standard. First we get the string length from the first 2 bytes
// then we fetch the string itself.
// Returns the decoded string, the total length of this section
// including the two bytes encoding the length, and a potential error.
func DecodeUTFString(toFetch []byte) (string, int, error) {
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

// Returns (MSB, LSB)
func getMSBandLSB(toEncode int) (byte, byte) {
	msb := byte(toEncode >> 8)
	lsb := byte(toEncode)
	return msb, lsb
}

func EncodeUTFString(toEncode string) ([]byte, int, error) {
	// If more than 16 bytes, 65535 = 2^16-1
	if len(toEncode) > 65535 {
		return []byte{}, 0, errors.New("error: String is too long to encode")
	}

	resultEncoding := make([]byte, len(toEncode)+2)
	resultEncoding[0], resultEncoding[1] = getMSBandLSB(len(toEncode))

	for pos, char := range toEncode {
		resultEncoding[2+pos] = byte(char)
	}

	return resultEncoding, len(toEncode) + 2, nil
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

func GetPacketType(packet []byte) byte {
	return packet[0] >> 4
}

var errZeroLengthPacketError = errors.New("error: Zero length packet read from byte pool")

// DecodePacket takes a byte array encoding a packet and returns
// (*Packet, PacketType, error)
func DecodePacket(packet []byte) (*Packet, byte, error) {

	if len(packet) == 0 {
		return nil, 0, errZeroLengthPacketError
	}

	packetType := GetPacketType(packet)

	var result *Packet
	var err error

	switch packetType {

	case CONNECT:
		result, err = DecodeConnect(packet[:])

	case CONNACK:
		result, err = DecodeCONNACK(packet[:])

	case SUBSCRIBE:
		result, err = DecodeSubscribe(packet[:])

	case PUBLISH:
		result, err = DecodePublish(packet[:])

	case PINGREQ:
		fmt.Println("Ping")
		result, err = DecodePing(packet[:])

	case DISCONNECT:
		result, err = DecodeDisconnect(packet[:])

	case SUBACK:
		result, err = DecodeSuback(packet[:])

	case UNSUBACK:
		result, err = DecodeUnsuback(packet[:])

	case UNSUBSCRIBE:
		result, err = DecodeUnsubscribe(packet[:])

	default:
		fmt.Println("Packet type not defined: ", packetType, " (", PacketTypeName(packetType), ")")
		return nil, 0, errPacketNotDefined
	}

	if err != nil {
		return nil, 0, err
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
		return nil, err
	}

	if fixedHeader.Type != 1 {
		err := fmt.Errorf("error: Incorrect packet type. Given type %v to connect ", fixedHeader.Type)
		return nil, err
	}

	resultPacket.ControlHeader = fixedHeader

	// Handle the variable length header
	varHeaderDecode := packet[fixedHeaderLen:]
	varHeader := ConnectVariableHeader{}

	protocolName, offset, err := DecodeUTFString(varHeaderDecode)
	if err != nil {
		return nil, err
	}
	varHeader.ProtocolName = protocolName

	protocolLevel, offset := varHeaderDecode[offset], offset+1
	varHeader.ProtocolLevel = protocolLevel
	flags, offset := varHeaderDecode[offset], offset+1

	varHeader.ConnectFlags = flags

	UsernameFlag := (flags>>7)&1 == 1
	PasswordFlag := (flags>>6)&1 == 1
	WillFlag := (flags>>2)&1 == 1
	// TODO: Think about these 3 flags
	// WillRetainFlag := (flags>>5)&1 == 1
	// WillQoS := (flags >> 3) & 3
	// CleanSession := (flags>>1)&1 == 1

	keepAlive := CombineMsbLsb(varHeaderDecode[offset], varHeaderDecode[offset+1])
	varHeader.KeepAlive = keepAlive
	offset += 2

	resultPacket.VariableLengthHeader = &varHeader

	// Now we've to decode the payload
	resultPayload := PacketPayload{}
	payloadDecode := varHeaderDecode[offset:]

	clientID, offset, err := DecodeUTFString(payloadDecode)
	if err != nil {
		return nil, err
	}
	resultPayload.ClientID = clientID

	if WillFlag {
		willTopic, addedOffset, err := DecodeUTFString(payloadDecode[offset:])
		if err != nil {
			return nil, err
		}
		offset += addedOffset

		willMessage, addedOffset, err := FetchBytes(payloadDecode[offset:])
		if err != nil {
			return nil, err
		}
		offset += addedOffset

		resultPayload.WillTopic = willTopic
		resultPayload.WillMessage = willMessage
	}

	if UsernameFlag {
		username, addedOffset, err := DecodeUTFString(payloadDecode[offset:])
		if err != nil {
			return nil, err
		}
		offset += addedOffset
		resultPayload.Username = username
	}

	if PasswordFlag {
		password, _, err := FetchBytes(payloadDecode[offset:])
		if err != nil {
			return nil, err
		}
		resultPayload.Password = &password
	}

	resultPacket.Payload = &resultPayload
	// If all goes well, we can return
	return resultPacket, nil
}

func DecodeCONNACK(packet []byte) (*Packet, error) {
	header, offset, err := DecodeFixedHeader(packet)
	if err != nil {
		return nil, err
	}

	varHeader := ConnackVariableHeader{}
	varHeader.ConnectAcknowledgementFlags = packet[offset]
	varHeader.ConnectReturnCode = packet[offset+1]

	return CombinePacketSections(header, &varHeader, nil), nil
}

func DecodeUnsubscribe(packet []byte) (*Packet, error) {

	resultPacket := &Packet{}
	fixedHeader, offset, err := DecodeFixedHeader(packet)
	resultPacket.ControlHeader = fixedHeader
	if err != nil {
		return nil, err
	}

	varHeader := UnsubscribeVariableHeader{}
	varHeader.PacketIdentifier = CombineMsbLsb(packet[offset], packet[offset+1])
	resultPacket.VariableLengthHeader = &varHeader
	offset += 2
	topics := make([]string, 0, 3)

	for offset < len(packet) {
		topic, addedOffset, err := DecodeUTFString(packet[offset:])
		if err != nil {
			return nil, errors.New("error: Malformed UTF string")
		}
		topics = append(topics, topic)
		offset += addedOffset
	}

	payload := PacketPayload{
		TopicList: ConvertStringsToTopicsWithQos(topics...),
	}
	resultPacket.Payload = &payload

	return resultPacket, nil

}

func DecodeSubscribe(packet []byte) (*Packet, error) {
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, offset, err := DecodeFixedHeader(packet)
	if err != nil {
		return nil, err
	}

	if fixedHeader.Flags != 2 {
		return nil, errors.New("error: Malformed Subscribe packet sent by client")
	}

	// Handle var header
	packetIdentifier := CombineMsbLsb(packet[offset], packet[offset+1])
	offset += 2
	varHeader := SubscribeVariableHeader{
		PacketIdentifier: packetIdentifier,
	}
	resultPacket.VariableLengthHeader = &varHeader

	// Get payload
	payload := PacketPayload{
		RawApplicationMessage: packet[offset:],
	}
	resultPacket.Payload = &payload

	return resultPacket, nil
}

func DecodeDisconnect(packet []byte) (*Packet, error) {

	resultPacket := &Packet{}
	resultPacket.ControlHeader = &ControlHeader{
		Type:            14,
		RemainingLength: 0,
		Flags:           0,
	}

	if packet[0]>>4 != 14 || packet[0]-14<<4 != 0 || packet[1] != 0 {

		return nil, errors.New("error: Incorrectly formed DISCONNECT packet")

	}

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
func DecodeVarLengthInt(toDecode []byte) (value int, length int, err error) {
	var multiplier int = 1
	for {
		encodedByte := toDecode[length]
		value += int((encodedByte & 127)) * multiplier
		multiplier *= 128

		if multiplier > 128*128*128 {
			return 0, 0, errMalformedInt
		}
		length += 1

		if encodedByte&128 == 0 {
			break
		}
	}
	return value, length, nil

}

func DecodePublish(packet []byte) (*Packet, error) {
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, offset, err := DecodeFixedHeader(packet)

	if err != nil {
		return nil, err
	}
	resultPacket.ControlHeader = fixedHeader

	// Handle the variable length header
	varHeader := PublishVariableHeader{}
	topicName, topicLen, err := DecodeUTFString(packet[offset:])
	varHeaderLen := topicLen

	// The qos is the second 2 bits of the flags
	qos := (fixedHeader.Flags&2 + fixedHeader.Flags&4) >> 1

	if qos == 1 || qos == 2 {
		packetIdentifier := CombineMsbLsb(packet[offset+topicLen], packet[offset+topicLen+1])
		varHeader.PacketIdentifier = packetIdentifier
		varHeaderLen += 2
	}

	if err != nil {
		return nil, err
	}

	varHeader.TopicFilter = topicName
	payloadLength := fixedHeader.RemainingLength - varHeaderLen
	offset = offset + varHeaderLen

	resultPacket.VariableLengthHeader = &varHeader

	var payload PacketPayload
	payload.RawApplicationMessage = make([]byte, payloadLength)
	copy(payload.RawApplicationMessage, packet[offset:offset+payloadLength])
	resultPacket.Payload = &payload

	return resultPacket, nil
}

func DecodePing(packet []byte) (*Packet, error) {
	resultPacket := &Packet{}
	// Handle the fixed length header
	fixedHeader, offset, err := DecodeFixedHeader(packet)
	if err != nil {
		return nil, err
	}
	resultPacket.ControlHeader = fixedHeader
	if offset != len(packet) {
		return nil, errInvalidLength
	}

	return resultPacket, nil

}

func DecodeSuback(packetArr []byte) (*Packet, error) {

	fixedHeader, offset, err := DecodeFixedHeader(packetArr)
	if err != nil {
		return nil, err
	}
	variableHeader := SubackVariableHeader{
		PacketIdentifier: CombineMsbLsb(packetArr[offset], packetArr[offset+1]),
	}
	payload := PacketPayload{
		RawApplicationMessage: packetArr[offset+2:],
	}

	resultPacket := Packet{
		ControlHeader:        fixedHeader,
		VariableLengthHeader: &variableHeader,
		Payload:              &payload,
	}
	return &resultPacket, nil

}

func DecodeUnsuback(packetArr []byte) (*Packet, error) {
	fixedHeader, offset, err := DecodeFixedHeader(packetArr)
	if err != nil {
		return nil, err
	}
	variableHeader := SubackVariableHeader{
		PacketIdentifier: CombineMsbLsb(packetArr[offset], packetArr[offset+1]),
	}

	if len(packetArr) != 4 {
		return nil, errors.New("error: Unsuback is incorrectly sized")
	}
	resultPacket := Packet{
		ControlHeader:        fixedHeader,
		VariableLengthHeader: &variableHeader,
	}
	return &resultPacket, nil

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
