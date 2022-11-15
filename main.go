package main

import (
	"fmt"

	tools "github.com/malachyoconnor/MQTT-GO/tools"
)

var (
	exampleHex = "18000000600edcea002206800000000000000000000000000000000100000000000000000000000000000001ec38075b4060bda33524ffff501827f6ab170000100c00044d5154540402003c0000"
)

func main() {
	fmt.Println("?")
	tools.Test()
	tools.DecodeHex(exampleHex)
}
