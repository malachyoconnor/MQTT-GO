package gobro

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"fmt"
)

type MessageHandler struct {
	AttachedInputChan  *chan client.ClientMessage
	AttachedOutputChan *chan []byte
}

func CreateMessageHandler(inputChanAddress *chan client.ClientMessage, outputChanAddress *chan []byte) MessageHandler {
	return MessageHandler{
		AttachedInputChan:  inputChanAddress,
		AttachedOutputChan: outputChanAddress,
	}
}

func (msgH *MessageHandler) Listen(server *Server) {

	for {
		clientMessage := <-(*msgH.AttachedInputChan)
		// clientID := *clientMessage.ClientID
		packet := clientMessage.Packet

		_, packetType, err := packets.DecodePacket(*packet)

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

		case packets.PUBLISH:
			// Check if the client exists by checking the client table
			// Then get the clients connected to that topic and send them
			// all a lovely packet
		case packets.SUBSCRIBE:
			// Check if the client exists
			// Then add them to the topic in the subscription table
		}
	}

}

func HandleConnect(connectPacket *packets.Packet, clientTable *map[string]*client.Client) {

	// if

}
