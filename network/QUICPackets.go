package network

// Long header packet types
const (
	LH_INITIAL byte = iota
	LH_RTT0
	LH_HANDSHAKE
	LH_RETRY
)

type LongHeaderPacket interface {
}

type LongHeader struct {
	HeaderForm                    bool   // set to True
	FixedBit                      bool   // set to True
	PacketType                    byte   // 2 bits
	TypeSpecificBits              byte   // 4 bits
	Version                       uint32 // 32 bits
	DestinationConnectionIDLength byte   // 8 bits
	DestinationConnectionID       []byte // 0..160 bits
	SourceConnectionIDLength      byte   // 8 bits
	SourceConnectionID            []byte // 0..160 bits
	// Then will follow with type specific values
}

type InitialPacket struct {
	*LongHeader // Note the packet type is zero
	// tokenLength       uint64 // Variable length int - don't need to store
	Token         []byte // Who knows how long
	Length        uint64 // Variable length int
	PacketNumber  uint32 // 8-32 bits
	PacketPayload []byte // 8-? bits
}

type RTT0Packet struct {
	LongHeader           // Note the packet type is 1
	length        uint64 // Variable length int
	packetNumber  uint32 // 8-32 bits
	packetPayload []byte // 8-? bits
}

type HandshakePacket struct {
	LongHeader           // Note the packet type is 2
	length        uint64 // Variable length int
	packetNumber  uint32 // 8-32 bits
	packetPayload []byte // 8-? bits
}

type RetryPacket struct {
	LongHeader               // Note the packet type is 3 AND the type specific bits are ignored
	retryToken        []byte // Who knows how long
	retryIntegrityTag []byte // 128 bits

}

// Short header packets

type ShortHeaderPacket struct {
	headerForm              bool   // set to False
	versionSpecificBits     byte   // 7 bits
	destinationConnectionID []byte // unknown len
}

type RTT1Packet struct {
	ShortHeaderPacket
	// Note the version specific bits are set to:
	// Fixed bit (1) Spin bit (1 (ignore this bit)) Reserved Bits (2) Key Phase (1) Packet Number Length (2)
	packetNumber  uint32 // 8-32 bits
	packetPayload []byte // 8-? bits
}
