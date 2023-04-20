package packets

// CombinePacketSections combines the control header, variable length header and payload into a single packet
// Note: VariableLengthHeader is an interface - so we don't pass a pointer
func CombinePacketSections(controlHeader *ControlHeader, varLengthHeader VariableLengthHeader, payload *PacketPayload) *Packet {
	resultPacket := Packet{}
	resultPacket.ControlHeader = controlHeader
	resultPacket.VariableLengthHeader = varLengthHeader
	resultPacket.Payload = payload
	return &resultPacket
}

// EncodeVarLengthInt encodes an integer into a variable length byte array as per the
// MQTT spec.
func EncodeVarLengthInt(initialInt int) []byte {
	result := make([]byte, 0)
	// Replicates a do while loop
	for intMoreThanZero := true; intMoreThanZero; intMoreThanZero = (initialInt > 0) {

		encodedByte := initialInt % 128
		initialInt /= 128

		// if there is more data to encode, set the top bit of this byte
		if initialInt > 0 {
			encodedByte |= 128
		}
		result = append(result, byte(encodedByte))
	}
	return result
}

// CreateConnACK creates a ConnACK packet
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

// CreateSubACK creates a SubACK packet
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

// CreateUnSuback creates an UnSuback packet
func CreateUnSuback(packetIdentifier int) []byte {
	result := make([]byte, 4)
	result[0] = UNSUBACK << 4
	result[1] = 2

	idMSB, idLSB := getMSBandLSB(packetIdentifier)
	result[2] = idMSB
	result[3] = idLSB
	return result
}
