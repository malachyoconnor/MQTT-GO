package packets

func EncodeConACK(packet *Packet) []byte {
	result := make([]byte, 4)
	result[0] = CONNACK << 4
	result[1] = byte(packet.ControlHeader.RemainingLength)
	result[2] = packet.VariableLengthHeader.(ConAckVariableHeader).connectAcknowledgeFlags
	result[3] = packet.VariableLengthHeader.(ConAckVariableHeader).connectReturnCode
	return result
}
