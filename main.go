package main

import (
	"fmt"

	tools "github.com/malachyoconnor/MQTT-GO/tools"
)

func main() {
	fmt.Println(tools.PacketType(3))

	fmt.Println(tools.CreateByte([]byte{1, 0, 0, 0, 0, 0, 1, 1}))
}
