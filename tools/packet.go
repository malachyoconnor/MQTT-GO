package tools

import "fmt"

type PacketType byte

const (
	RESERVED PacketType = iota
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
	AUTH
)

func (packet PacketType) toByte() byte {
	return byte(packet)
}

var ConnectStruct struct {
	ControlHeader        *ControlHeader
	VariableLengthHeader *VariableLengthHeader
	Payload              *[]byte
}

type VariableLengthHeader interface{}

type ControlHeader struct {
	Type        PacketType
	FixedLength int
	// Flags
	Qos    byte
	Dup    bool
	Retain bool
}

type ConnectVariableLengthHeader struct {
	a int
}

func test() {
	var x VariableLengthHeader = ConnectVariableLengthHeader{a: 5}
	fmt.Printf(x)
}
