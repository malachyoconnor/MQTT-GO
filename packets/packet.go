package packets

const (
	RESERVED byte = iota
	CONNECT       // Done
	CONNACK       // Done
	PUBLISH       // Done
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE   // Done
	SUBACK      // Done
	UNSUBSCRIBE // Done
	UNSUBACK    //Done
	PINGREQ     // Done
	PINGRESP
	DISCONNECT // Done
	AUTH       = 15
)

// CONNECT
// CONNACK
// SUBSCRIBE
// PUBLISH
// PINGREQ
// DISCONNECT
// SUBACK
// UNSUBACK
// UNSUBSCRIBE

func PacketTypeName(packetType byte) string {
	if packetType > 15 {
		return "UNDEFINED"
	}
	return []string{
		"RESERVED", "CONNECT", "CONNACK", "PUBLISH", "PUBACK",
		"PUBREC", "PUBREL", "PUBCOMP", "SUBSCRIBE", "SUBACK", "UNSUBSCRIBE",
		"UNSUBACK", "PINGREQ", "PINGRESP", "DISCONNECT", "AUTH",
	}[packetType]
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
func (*PublishVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*ConnectVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*SubscribeVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*ConnackVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*SubackVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*UnsubscribeVariableHeader) SafetyFunc() {
	//Safety func to ensure passing pointers to variable headers - remove once finished development
}

type ControlHeader struct {
	RemainingLength int
	Type            byte
	Flags           byte
}

type PublishVariableHeader struct {
	TopicFilter      string
	PacketIdentifier int
}

type ConnectVariableHeader struct {
	ProtocolName  string
	KeepAlive     int
	ConnectFlags  byte
	ProtocolLevel byte
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

type SubackVariableHeader struct {
	PacketIdentifier int
}

type UnsubscribeVariableHeader struct {
	PacketIdentifier int
}

type SubscribeVariableHeader struct {
	PacketIdentifier int
}

type PacketPayload struct {
	ClientID              string
	WillTopic             string
	WillMessage           []byte
	Username              string
	Password              *[]byte
	TopicList             []TopicWithQoS
	RawApplicationMessage []byte
}

type TopicWithQoS struct {
	Topic string
	QoS   byte
}
