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

func isValidControlType(controlType byte) bool {
	return ((RESERVED < controlType) && (controlType <= AUTH))
}

type Packet struct {
	ControlHeader        *ControlHeader
	VariableLengthHeader *VariableLengthHeader
	Payload              *[]byte
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

type ConnectVariableLengthHeader struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  byte
	KeepAlive     int
}
