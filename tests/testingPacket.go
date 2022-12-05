package tests

import (
	"fmt"
	"os"

	"github.com/malachyoconnor/MQTT-GO/packets"
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
