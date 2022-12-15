package packets

func CreateConnACK(cleanSession bool, returnCode byte) Packet {
	resultPacket := Packet{}
	resultPacket.ControlHeader = ControlHeader{
		Type:            CONNACK,
		RemainingLength: 2,
	}

	var connectAcknowledgeFlags byte
	if cleanSession {
		connectAcknowledgeFlags = 1
	}

	variableHeader := ConAckVariableHeader{
		connectAcknowledgeFlags: connectAcknowledgeFlags,
		connectReturnCode:       returnCode,
	}

	resultPacket.VariableLengthHeader = variableHeader
	return resultPacket
}
