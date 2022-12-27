package packets

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
