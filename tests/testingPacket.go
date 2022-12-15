package tests

import (
	"fmt"
	"os"

	"MQTT-GO/packets"
)

func TestConnect() {

	examplePacket, err := os.ReadFile("./tests/exampleConnect")

	fmt.Println(err)

	if err != nil {
		return
	}

	fmt.Println(examplePacket)
	resultPacket, err := packets.DecodeConnect(examplePacket)

	fmt.Println(err)

	fmt.Println(resultPacket)

}

// Example connect packet

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
	0, 3, // Length of the client identifier
	'm', 'a', 'l',
}
