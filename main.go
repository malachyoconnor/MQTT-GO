package main

import (
	"encoding/json"
	"fmt"

	packets "github.com/malachyoconnor/MQTT-GO/packets"
)

func print(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

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

func main() {

	// fmt.Println(packets.FetchUTFString(connectPacket[2:]))
	packet, err := packets.DecodeConnect(connectPacket[:])
	if err != nil {
		fmt.Println(err)
	} else {
		print(packet)
	}

}
