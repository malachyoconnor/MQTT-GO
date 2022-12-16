package gobro

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"fmt"
	"strings"
)

type MessageHandler struct {
	AttachedInputChan  *chan client.ClientMessage
	AttachedOutputChan *chan client.ClientMessage
}

func CreateMessageHandler(inputChanAddress *chan client.ClientMessage, outputChanAddress *chan client.ClientMessage) MessageHandler {
	return MessageHandler{
		AttachedInputChan:  inputChanAddress,
		AttachedOutputChan: outputChanAddress,
	}
}

func (msgH *MessageHandler) Listen(server *Server) {

	clientTable := *server.clientTable

	for {
		clientMessage := <-(*msgH.AttachedInputChan)
		clientID := client.ClientID(*clientMessage.ClientID)
		packetArray := clientMessage.Packet
		clientConnection := *(clientMessage.ClientConnection)

		packet, packetType, err := packets.DecodePacket(*packetArray)

		if err != nil {
			fmt.Println(err)
			continue
		}

		// General case for if the client doesn't exist IF NOT A CONNECT packet
		if packetType != packets.CONNECT {

			if _, found := clientTable[clientID]; !found {
				fmt.Println("Client not in the client table sending messages, disconnecting.")
				clientConnection.Close()
				continue
			}

		}

		switch packetType {
		case packets.CONNECT:
			// Query the client table to check if the client exists
			// if not slap it in there - then send connack

			// This should disconnect them if they're already connected !!
			createClient(&clientTable, &clientMessage)
			// Check if the reserved flag is zero, if not disconnect them
			// Finally send out a CONACK [X]
			connackPacket := packets.CreateConnACK(false, 0)
			connackArray := packets.EncodeConACK(&connackPacket)
			clientMsg := client.CreateClientMessage(&clientID, &clientConnection, &connackArray)
			(*server.outputChan) <- clientMsg

		case packets.PUBLISH:
			var stringBuilder strings.Builder
			stringBuilder.Write(packet.Payload.ApplicationMessage)
			fmt.Println("Received request to publish:", stringBuilder.String())

			// Get the clients connected to that topic and send them
			// all a lovely packet
		case packets.SUBSCRIBE:
			// Check if the client exists
			// Then add them to the topic in the subscription table

			fmt.Println("Handling subscribe")

		case packets.DISCONNECT:
			// Close the client TCP connection.
			// Remove the packet from the client list
			clientConnection.Close()
			delete(clientTable, clientID)
		}
	}

}

func createClient(clientTable *client.ClientTable, clientMsg *client.ClientMessage) *client.Client {

	if client, found := (*clientTable)[*clientMsg.ClientID]; found {
		return client
	}

	client := &client.Client{
		ClientIdentifier: client.ClientID(*clientMsg.ClientID),
		TCPConnection:    *clientMsg.ClientConnection,
		Topics:           make([]string, 10),
	}

	(*clientTable)[*clientMsg.ClientID] = client

	return client
}
