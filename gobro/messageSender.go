package gobro

import "MQTT-GO/packets"

type MessageSender struct {
	outputPool *packets.BytePool
}

func CreateMessageSender(outputPool *packets.BytePool) MessageSender {
	return MessageSender{
		outputPool: outputPool,
	}
}

func (MessageSender) SendMessages(server *Server) {

}
