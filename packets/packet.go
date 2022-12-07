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

func packetTypeName(packetType byte) string {
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
	Type        byte
	FixedLength int
	// Flags
	Qos    byte
	Dup    bool
	Retain bool
}

type PublishVariableLengthHeader struct {
	TopicName        string
	PacketIdentifier int
}

type ConnectVariableLengthHeader struct {
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

type PacketPayload struct {
	clientId    string
	willTopic   string
	willMessage []byte
	username    string
	password    string
	topicName   string

	ApplicationMessage []byte
}
