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
	// Flags
	Qos    byte
	Dup    bool
	Retain bool
}

type PublishVariableHeader struct {
	TopicName        string
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

type PacketPayload struct {
	ClientId    string
	WillTopic   string
	WillMessage []byte
	Username    string
	Password    string
	TopicName   string

	ApplicationMessage []byte
}

type ConAckVariableHeader struct {
	connectAcknowledgeFlags byte
	connectReturnCode       byte
}

func (packet *Packet) GetVarHeader() VariableLengthHeader {
	// TODO: Finish these

	varLengthHeader := packet.VariableLengthHeader

	switch packet.ControlHeader.Type {

	case CONNECT:
		return varLengthHeader.(ConnectVariableHeader)
	case CONNACK:
		return varLengthHeader.(ConAckVariableHeader)
	}
	return nil
}
