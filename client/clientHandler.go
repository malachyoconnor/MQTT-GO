package client

import (
	"MQTT-GO/packets"
	"fmt"
	"net"
)

func ClientHandler(connection *net.Conn, packetPool *packets.BytePool) {
	for i := 0; i < 10; i++ {

		// FIXME: Consider changing this and optimizing for speed!
		// 	      Try and find the best value to initialize this to.
		buffer := make([]byte, 300)
		packetLen, err := (*connection).Read(buffer)

		packet := make([]byte, packetLen)
		copy(packet, buffer[:packetLen])

		if err != nil {
			fmt.Println(err)
			continue
		}

		packetPool.Put(packet)
	}
}
