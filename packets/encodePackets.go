package packets

func EncodeConACK(packet *Packet) [4]byte {
	var result [4]byte
	result[0] = CONNECT << 4
	result[1] = byte(packet.ControlHeader.RemainingLength)
	result[2] = packet.VariableLengthHeader.(ConAckVariableLengthHeader).connectAcknowledgeFlags
	result[3] = packet.VariableLengthHeader.(ConAckVariableLengthHeader).connectReturnCode
	return result
}
