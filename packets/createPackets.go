package packets

import "errors"

// VariableLengthHeader is an interface - so we don't pass a pointer
func CombinePacketSections(controlHeader *ControlHeader, varLengthHeader VariableLengthHeader, payload *PacketPayload) *Packet {
	resultPacket := Packet{}
	resultPacket.ControlHeader = controlHeader
	resultPacket.VariableLengthHeader = varLengthHeader
	resultPacket.Payload = payload
	return &resultPacket
}

func EncodeFixedHeader(fixedHeader ControlHeader) []byte {

	result := make([]byte, 1)
	result[0] = (fixedHeader.Type << 4) | (fixedHeader.Flags)
	result = append(result, EncodeVarLengthInt(fixedHeader.RemainingLength)...)

	return result
}

func EncodeVarLengthInt(initialInt int) []byte {
	result := make([]byte, 0)
	// Replicates a do while loop
	for intMoreThanZero := true; intMoreThanZero; intMoreThanZero = (initialInt > 0) {

		encodedByte := initialInt % 128
		initialInt = initialInt / 128

		// if there is more data to encode, set the top bit of this byte
		if initialInt > 0 {
			encodedByte = encodedByte | 128
		}
		result = append(result, byte(encodedByte))
	}
	return result
}

func CreateConnACK(cleanSession bool, returnCode byte) []byte {
	result := make([]byte, 4)
	result[0] = CONNACK << 4
	result[1] = 2 // Remaining length
	var connectAcknowledgeFlags byte
	if cleanSession {
		connectAcknowledgeFlags = 1
	}
	result[2] = connectAcknowledgeFlags
	result[3] = returnCode
	return result
}

func CreateSubACK(packetIdentifier int, returnCodes []byte) []byte {

	result := make([]byte, 4+byte(len(returnCodes)))
	result[0] = SUBACK << 4
	result[1] = 2 + byte(len(returnCodes))

	idMSB, idLSB := getMSBandLSB(packetIdentifier)
	result[2] = idMSB
	result[3] = idLSB

	for i, code := range returnCodes {
		result[4+i] = code
	}

	return result
}

// TODO: We discard errors in this function...
func CreateConnect(packet *Packet) (*[]byte, error) {
	if packet.ControlHeader.Type != CONNECT {
		panic("Tried to create a connect packet from a non-connect packet")
	}
	// Need to do control header last because only at the end can we know the value of the remaining length field

	varLengthHeader := packet.VariableLengthHeader.(*ConnectVariableHeader)
	// TODO: Random choice of 30 here - could be improved with some looking into, same for the Payload.
	resultVarHeader := make([]byte, 0, 30)
	protocolNameArr, _, _ := EncodeUTFString("MQTT")

	// Version 3.1.1 has a protocol version of 4
	protocol := byte(4)
	connectFlags := varLengthHeader.ConnectFlags
	keepAliveMsb, keepAliveLsb := getMSBandLSB(varLengthHeader.KeepAlive)

	resultVarHeader = append(resultVarHeader, protocolNameArr...)
	resultVarHeader = append(resultVarHeader, protocol, connectFlags, keepAliveMsb, keepAliveLsb)

	payload := packet.Payload
	resultPayload := make([]byte, 0, 30)
	// Client Identifier
	clientIdentifier, _, err := EncodeUTFString(payload.ClientID)
	if err != nil {
		return nil, errors.New("error: clientID not provided")
	}

	resultPayload = append(resultPayload, clientIdentifier...)

	// Will Topic & Will Message (If the will flag is set to 1)
	if (varLengthHeader.ConnectFlags & 4) > 0 {
		if payload.WillTopic != "" || payload.WillMessage == nil {
			return nil, errors.New("error: Will metadata not provided")
		}
		willTopic, _, _ := EncodeUTFString(payload.WillTopic)
		willMessage, _, _ := EncodeUTFString(payload.WillTopic)
		resultPayload = append(resultPayload, willTopic...)
		resultPayload = append(resultPayload, willMessage...)
	}
	// User Name
	if (varLengthHeader.ConnectFlags & 128) > 0 {
		if payload.Username == "" {
			return nil, errors.New("error: Username metadata not provided")
		}
		username, _, _ := EncodeUTFString(payload.Username)
		resultPayload = append(resultPayload, username...)
	}

	// Password
	if (varLengthHeader.ConnectFlags & 64) > 0 {
		if payload.Password == nil {
			return nil, errors.New("error: Password metadata not provided")
		}
		passwordLenMSB, passwordLenLSB := getMSBandLSB(len(*payload.Password))
		resultPayload = append(resultPayload, passwordLenMSB, passwordLenLSB)
		resultPayload = append(resultPayload, *payload.Password...)
	}

	packet.ControlHeader.RemainingLength = len(resultPayload) + len(resultVarHeader)
	resultControlHeader := EncodeFixedHeader(*packet.ControlHeader)

	resultPacket := make([]byte, 0, len(resultControlHeader)+len(resultVarHeader)+len(resultPayload))
	resultPacket = append(resultPacket, resultControlHeader...)
	resultPacket = append(resultPacket, resultVarHeader...)
	resultPacket = append(resultPacket, resultPayload...)

	return &resultPacket, nil
}
