package packets

import "errors"

func CombineEncodedPacketSections(controlHeader []byte, varLengthHeader []byte, payload []byte) []byte {
	resultPacket := make([]byte, 0, len(controlHeader)+len(varLengthHeader)+len(payload))
	resultPacket = append(resultPacket, controlHeader...)
	resultPacket = append(resultPacket, varLengthHeader...)
	resultPacket = append(resultPacket, payload...)
	return resultPacket
}

func EncodeFixedHeader(fixedHeader ControlHeader) []byte {

	result := make([]byte, 1)
	result[0] = (fixedHeader.Type << 4) | (fixedHeader.Flags)
	result = append(result, EncodeVarLengthInt(fixedHeader.RemainingLength)...)

	return result
}

// TODO: We discard errors in this function...
func EncodeConnect(packet *Packet) ([]byte, error) {
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

	return CombineEncodedPacketSections(resultControlHeader, resultVarHeader, resultPayload), nil
}

func EncodePublish(packet *Packet) ([]byte, error) {
	if packet.ControlHeader.Type != PUBLISH {
		panic("Error create publish passed non-publish packet")
	}

	resultVarHeader := make([]byte, 0, 30)
	varLenHeader := packet.VariableLengthHeader.(*PublishVariableHeader)
	topicName, _, err := EncodeUTFString(varLenHeader.TopicFilter)
	resultVarHeader = append(resultVarHeader, topicName...)
	if err != nil {
		return nil, err
	}

	qos := packet.ControlHeader.Flags & 6
	if qos > 0 {
		packetIdMSB, packetIdLSB := getMSBandLSB(varLenHeader.PacketIdentifier)
		resultVarHeader = append(resultVarHeader, packetIdMSB, packetIdLSB)
	}

	resultPayload := packet.Payload.ApplicationMessage

	packet.ControlHeader.RemainingLength = len(resultPayload) + len(resultVarHeader)
	resultControlHeader := EncodeFixedHeader(*packet.ControlHeader)

	return CombineEncodedPacketSections(resultControlHeader, resultVarHeader, resultPayload), nil
}

func EncodeSubscribe(packet *Packet) ([]byte, error) {

	if packet.ControlHeader.Type != SUBSCRIBE {
		panic("Error create publish passed non-publish packet")
	}
	packetIdentifier := packet.VariableLengthHeader.(*SubscribeVariableHeader).PacketIdentifier
	resultVarHeader := make([]byte, 2)
	resultVarHeader[0], resultVarHeader[1] = getMSBandLSB(packetIdentifier)
	resultPayload := packet.Payload.ApplicationMessage
	packet.ControlHeader.RemainingLength = len(resultVarHeader) + len(resultPayload)
	resultControlHeader := EncodeFixedHeader(*packet.ControlHeader)

	return CombineEncodedPacketSections(resultControlHeader, resultVarHeader, resultPayload), nil

}

func EncodeUnsubscribe(packet *Packet) ([]byte, error) {
	if packet.ControlHeader.Type != UNSUBSCRIBE {
		panic("Error encode unsubscribe passed non-unsubscribe packet")
	}
	packetIdentifier := packet.VariableLengthHeader.(*UnsubscribeVariableHeader).PacketIdentifier
	resultVarHeader := make([]byte, 2)
	resultVarHeader[0], resultVarHeader[1] = getMSBandLSB(packetIdentifier)
	resultPayload := make([]byte, 0, len(packet.Payload.TopicList))
	for _, topic := range packet.Payload.TopicList {
		encodedTopic, _, err := EncodeUTFString(topic)
		if err != nil {
			return nil, nil
		}
		resultPayload = append(resultPayload, encodedTopic...)
	}
	packet.ControlHeader.RemainingLength = len(resultVarHeader) + len(resultPayload)
	resultControlHeader := EncodeFixedHeader(*packet.ControlHeader)

	return CombineEncodedPacketSections(resultControlHeader, resultVarHeader, resultPayload), nil

}
