// Package packets does all of the work of creating, encoding and decoding MQTT packets.
// It also contains the structs for the packets.
package packets

// This stores the MQTT packet types and the associated IDs
// Its passed in the control header of the packet
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

// PacketTypeName returns the name of a given packet type as a string
func PacketTypeName(packetType byte) string {
	if packetType > AUTH {
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

// Packet is the main struct for an MQTT packet
// It consists of a ControlHeader, a VariableLengthHeader and a Payload
type Packet struct {
	ControlHeader *ControlHeader
	// VariableLengthHeader is an INTERFACE so it contains a pointer to the struct anyway.
	// Pointers to interfaces are mostly useless
	VariableLengthHeader VariableLengthHeader
	Payload              *PacketPayload
}

// VariableLengthHeader is an interface for the variable length header of a packet
// There are multiple types of variable length header, so we use an interface
type VariableLengthHeader interface {
	SafetyFunc()
}

// This ensures we only pass around POINTERS to our variable header structs
// Otherwise we could be passing the structs itself and not realise it
func (*PublishVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*ConnectVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*SubscribeVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*ConnackVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*SubackVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}
func (*UnsubscribeVariableHeader) SafetyFunc() {
	// Safety func to ensure passing pointers to variable headers - remove once finished development
}

// ControlHeader is the header of the packet that contains the packet type and flags
type ControlHeader struct {
	RemainingLength int
	Type            byte
	Flags           byte
}

// PublishVariableHeader is the variable header for a publish packet
type PublishVariableHeader struct {
	TopicFilter      string
	PacketIdentifier int
}

// ConnectVariableHeader is the variable header for a connect packet
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

// PacketPayload is the type used for all payloads of packets
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
