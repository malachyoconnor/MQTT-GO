package gobro

import (
	"MQTT-GO/packets"
	"fmt"
)

type MessageHandler struct {
	AttachedBytePool *BytePool
}

func CreateMessageHandler(bytePoolAddress *BytePool) *MessageHandler {
	return &MessageHandler{
		AttachedBytePool: bytePoolAddress,
	}
}

func (msgH *MessageHandler) Listen() {

	for {
		newPacket := msgH.AttachedBytePool.Get()

		result, err := packets.DecodePacket(newPacket)
		if err != nil {
			fmt.Println(err)
			continue
		}

		packets.PrintPacket(result)
	}

}
