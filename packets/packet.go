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
	if packetType > 15 || packetType < 0 {
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
	ConnectFlags  byte
	KeepAlive     int

	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQoS        byte
	WillFlag       bool
	CleanSession   bool
}

type SubscribeVariableHeader struct {
	PacketIdentifier int
}

type PacketPayload struct {
	ClientId           string
	WillTopic          string
	WillMessage        []byte
	Username           string
	Password           string
	TopicName          string
	ApplicationMessage []byte
}

type ConAckVariableHeader struct {
	connectAcknowledgeFlags byte
	connectReturnCode       byte
}
