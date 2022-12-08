package packets

func CreateConACK(cleanSession bool, returnCode byte) Packet {
	resultPacket := Packet{}
	resultPacket.ControlHeader = ControlHeader{
		Type:            CONNECT,
		RemainingLength: 2,
	}

	var connectAcknowledgeFlags byte
	if cleanSession {
		connectAcknowledgeFlags = 1
	}

	variableHeader := ConAckVariableLengthHeader{
		connectAcknowledgeFlags: connectAcknowledgeFlags,
		connectReturnCode:       returnCode,
	}

	resultPacket.VariableLengthHeader = variableHeader
	return resultPacket
}
