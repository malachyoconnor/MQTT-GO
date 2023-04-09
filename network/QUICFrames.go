package network

const (
	PADDING byte = iota
	PING
	ACK
	RESET_STREAM
	STOP_SENDING
	CRYPTO
	NEW_TOKEN
	STREAM
	MAX_DATA
	MAX_STREAM_DATA
	MAX_STREAMS
	DATA_BLOCKED
	STREAM_DATA_BLOCKED
	STREAMS_BLOCKED
	NEW_CONNECTION_ID
	RETIRE_CONNECTION_ID
	PATH_CHALLENGE
	PATH_RESPONSE
	CONNECTION_CLOSE
	HANDSHAKE_DONE
)

type Frame interface{}

type PINGFrame struct {
	frameType uint64 // var-length int - should be set to 1
}

type ACKFrame struct {
	frameType           uint64     // Should be set to 2 or 3
	largestAcknowledged uint64     // Variable length int
	ackDelay            uint64     // Variable length int
	ackrangeCount       uint64     // Variable length int
	firstAckRange       uint64     // Variable length int
	ackRanges           []ACKRange // Multiple ack ranges
	// Not including the ECN counts
}

type ACKRange struct {
	gap            uint64 // Var length int
	ackRangeLength uint64 // Var length int
}

type RESET_STREAMFrame struct {
	frameType                    uint64 // var length. Should be set to 4
	streamID                     uint64 // var length
	applicationProtocolErrorCode uint64 // var length
	finalSize                    uint64 // var length
}

type STOP_SENDINGFrame struct {
	frameType                    uint64 // var length. Should be set to 5
	streamID                     uint64 // var length
	applicationProtocolErrorCode uint64 // var length
}

type CRYPTOFrame struct {
	frameType  uint64 // var length. Should be set to 6
	offset     uint64 // var length
	length     uint64 // var length
	cryptoData []byte // Who knows how long
}

type NEW_TOKENFrame struct {
	frameType   uint64 // var length. should be set to 7
	tokenLength uint64 // var length
	token       []byte // Who knows how long
}

type STREAMFrame struct {
	// frameType takes the form 0b0001XXX. The three low-order bits determine
	// the fields present in the frame.
	// Bit 4 says if there is an offset field present
	// Bit 2 says if there is a length field present
	// Bit 1 says if this is the end of the stream
	frameType  uint64
	streamID   uint64 // var
	streamData []byte // As mentioned, CAN start with the offset followed by the length - both var length
}

// Might not need
type MAX_DATAFrame struct {
	frameType   uint64 // var length int. Set to 10
	maximumData uint64 //var length
}

// Might not need
type MAX_STREAM_DATAFrame struct {
	frameType         uint64 // var length int. Set to 11
	streamID          uint64 // var length
	maximumStreamData uint64 // var length
}

// Not implementing max streams

type DATA_BLOCKEDFrame struct {
	frameType   uint64 // var length int. Set to 14
	maximumData uint64 // var length
}

type CONNECTION_CLOSEFrame struct {
	frameType          uint64 // var length. Set to 28 for signal error at the quic level. 29 for app level error
	errorCode          uint64 // var length
	errorframeType     byte   // Just set to 0
	reasonPhraseLength uint64 // var length
	reasonPhrase       []byte // Reason for the error - can be empty
}

type HANDSHAKE_DONEFrame struct {
	frameType uint64 // var length. Set to 30
}
