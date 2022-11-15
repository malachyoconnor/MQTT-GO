package tools

import (
	h "encoding/hex"
	"fmt"
)

var connectPacket = []byte{
	// Fixed header
	(CONNECT.toByte() << 4), // Define the control packet type, shift as the first nibble is type and the second is reserved
	15,                      // Define the fixed length of the rest of the packet
	// Variable length header
	// - Protocol Name
	0, 4, // Length of the protocol name
	'M', 'Q', 'T', 'T',
	// - Protocol Version
	5,
	// - Connect Flags
	CreateByteInline([]byte{0, 0, 0, 0, 0, 0, 0, 0}), // Only the will flag is set to 1
	// - KeepAlive
	0, 10,
	// - Properties
	0, // Length 0, no properties set
	// Payload
	3, // Length of the client identifier
	'm', 'a', 'l',
}

func CreateByteInline(input_binary []byte) byte {
	res, _ := CreateByte(input_binary)
	return res
}

// Take an array of 0s and 1s and return the byte representation
func CreateByte(input_binary []byte) (byte, bool) {
	if len(input_binary) > 8 {
		return 0, false
	}
	var result byte = 0
	for i, v := range input_binary {
		if !(v == 0 || v == 1) {
			return 0, false
		}
		result += (v << (7 - i))
	}
	return result, true
}

func DecodeHex(hex string) {
	result, err := h.DecodeString(hex)

	for x := range result {
		fmt.Printf("%b\n", x)
	}

	fmt.Println(result, err)

	fmt.Println("Got here")
}
