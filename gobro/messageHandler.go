package gobro

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"fmt"
)

type MessageHandler struct {
	AttachedInputPool  *packets.BytePool
	AttachedOutputPool *packets.BytePool
}

func CreateMessageHandler(inputPoolAddress *packets.BytePool, outputPoolAddress *packets.BytePool) MessageHandler {
	return MessageHandler{
		AttachedInputPool:  inputPoolAddress,
		AttachedOutputPool: outputPoolAddress,
	}
}

func (msgH *MessageHandler) Listen(server *Server) {

	for {
		newPacket := msgH.AttachedInputPool.Get()

		decodedPacket, packetType, err := packets.DecodePacket(newPacket)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch packetType {
		case packets.CONNECT:
			// Query the client table to check if the client exists
			// if not slap it in there - then send connack
			// If it IS in there, then we disconnect them
			// If the connect packet is empty we give our own

			// Check if the reserved flag is zero, if not disconnect them
			// Finally send out a CONACK [X]

			conACK := packets.CreateConACK(true, 0)
			resultToSend := packets.EncodeConACK(&conACK)

			msgH.AttachedOutputPool.Put(resultToSend[:])

		case packets.PUBLISH:
			// Check if the client exists by checking the client table
			// Then get the clients connected to that topic and send them
			// all a lovely packet
		case packets.SUBSCRIBE:
			// Check if the client exists
			// Then add them to the topic in the subscription table
		}
		packets.PrintPacket(decodedPacket)
	}

}

func HandleConnect(connectPacket *packets.Packet, clientTable *map[string]*client.Client) {

	// if

}
