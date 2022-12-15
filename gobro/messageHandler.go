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

	clientTable := *server.clientTable

	for {
		clientMessage := <-(*msgH.AttachedInputChan)
		clientID := client.ClientID(*clientMessage.ClientID)
		packet := clientMessage.Packet
		clientConnection := *(clientMessage.ClientConnection)

		_, packetType, err := packets.DecodePacket(*packet)

		if err != nil {
			fmt.Println(err)
			continue
		}

		switch packetType {
		case packets.CONNECT:
			// Query the client table to check if the client exists
			// if not slap it in there - then send connack

			createClient(&clientTable, &clientMessage)

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

func HandleConnect(connectPacket *packets.Packet, clientTable *map[string]*client.Client) {

	// if

}
