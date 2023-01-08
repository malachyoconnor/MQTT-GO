package packets

const (
	RESERVED byte = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH = 15
)

func PacketTypeName(packetType byte) string {
	if packetType > 15 {
		return "UNDEFINED"
	}
	return []string{"RESERVED", "CONNECT", "CONNACK", "PUBLISH", "PUBACK",
		"PUBREC", "PUBREL", "PUBCOMP", "SUBSCRIBE", "SUBACK", "UNSUBSCRIBE",
		"UNSUBACK", "PINGREQ", "PINGRESP", "DISCONNECT", "AUTH"}[packetType]

}

func isValidControlType(controlType byte) bool {
	return ((RESERVED < controlType) && (controlType <= AUTH))
}

type Packet struct {
	ControlHeader        ControlHeader
	VariableLengthHeader VariableLengthHeader
	Payload              PacketPayload
}

type VariableLengthHeader interface{}

type ControlHeader struct {
	Type            byte
	RemainingLength int
	Flags           byte
}

type PublishVariableHeader struct {
	TopicFilter      string
	PacketIdentifier int
}

type ConnectVariableHeader struct {
	ProtocolName  string
	ProtocolLevel byte
	KeepAlive     int
	ConnectFlags  byte
	// The flags are (IN INCREASING ORDER):

	// Reserved (1 bit) set to 0
	// Clean Session (1 bit)
	// Will Flag (1 bit)
	// Will QoS (2 bits)
	// Will Retain (1 bit)
	// Password Flag (1 bit)
	// User Name Flag (1 bit)
}

type SubscribeVariableHeader struct {
	PacketIdentifier int
}

// TODO: change the byte lists to pointers to byte lists
type PacketPayload struct {
	ClientID           string
	WillTopic          string
	WillMessage        []byte
	Username           string
	Password           *[]byte
	TopicName          string
	ApplicationMessage []byte
}

//lint:ignore U1000 We might use this in the future
type ConAckVariableHeader struct {
	connectAcknowledgeFlags byte
	connectReturnCode       byte
}
