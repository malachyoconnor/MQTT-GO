package network

import (
	"MQTT-GO/structures"
	"fmt"
	"net"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func RunTest() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8000})

	panicErr(err)
	defer conn.Close()

	buffer := make([]byte, 300)
	_, err = conn.Read(buffer)
	panicErr(err)

	res, err := decodeLongHeaderPacket(buffer)
	x := InitialPacket(*res.(*InitialPacket))
	panicErr(err)

	structures.PrintInterface(x)

	_, err = decodeFrame(x.PacketPayload)
	panicErr(err)

	fmt.Println()

}
