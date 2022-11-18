package main

import (
	"fmt"

	packets "github.com/malachyoconnor/MQTT-GO/packets"
)

var connectPacket = []byte{
	// Fixed header
	(packets.CONNECT << 4), // Define the control packet type, shift as the first nibble is type and the second is reserved
	15,                     // Define the fixed length of the rest of the packet
	// Variable length header
	// - Protocol Name
	0, 4, // Length of the protocol name
	'M', 'Q', 'T', 'T',
	// - Protocol Version
	5,
	// - Connect Flags
	packets.CreateByteInline([]byte{0, 0, 0, 0, 0, 0, 0, 0}), // Only the will flag is set to 1
	// - KeepAlive
	0, 10,
	// - Properties
	0, // Length 0, no properties set
	// Payload
	3, // Length of the client identifier
	'm', 'a', 'l',
}

func main() {
	fmt.Println(packets.CreateByte([]byte{1, 0, 0, 0, 0, 0, 0, 0}))

	// fmt.Println(tools.CreateByte([]byte{1, 0, 0, 0, 0, 0, 1, 1}))

	packets.DecodeConnect(connectPacket[:])
}
