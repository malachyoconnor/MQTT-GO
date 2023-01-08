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

// Why is this not pointers to the items????
type Packet struct {
	ControlHeader *ControlHeader
	// VariableLengthHeader is an INTERFACE so it contains a pointer to the struct anyway.
	// Pointers to interfaces are mostly useless
	VariableLengthHeader VariableLengthHeader
	Payload              *PacketPayload
}

type VariableLengthHeader interface {
	SafetyFunc()
}

// This ensures we only pass around POINTERS to our variable header structs
// Otherwise we could be passing the structs itself and not realise it
func (*PublishVariableHeader) SafetyFunc()   {}
func (*ConnectVariableHeader) SafetyFunc()   {}
func (*SubscribeVariableHeader) SafetyFunc() {}
func (*ConnackVariableHeader) SafetyFunc()   {}

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

type ConnackVariableHeader struct {
	ConnectAcknowledgementFlags byte
	ConnectReturnCode           byte
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
