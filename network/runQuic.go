package network

import (
	"MQTT-GO/structures"
	"fmt"
	"net"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func RunQuic() {
	fmt.Println("Listening for QUIC packets")
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8000})
	checkError(err)
	defer conn.Close()

	buffer := make([]byte, 4096)
	packetLen, err := conn.Read(buffer)
	checkError(err)
	fmt.Println("Read a packet")
	res, err := decodeLongHeaderPacket(buffer[:packetLen])
	initialPacket := res.(*InitialPacket)
	checkError(err)

	initialPacket.PacketPayload = []byte{0}
	structures.PrintInterface(initialPacket)

}
