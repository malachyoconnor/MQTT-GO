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
			connackArray := packets.CreateConnACK(false, 0)
			clientMsg := client.CreateClientMessage(&clientID, &clientConnection, &connackArray)
			(*server.outputChan) <- clientMsg

		case packets.PUBLISH:
			var stringBuilder strings.Builder
			stringBuilder.Write(packet.Payload.ApplicationMessage)
			fmt.Println("Received request to publish:", stringBuilder.String())

			
			varHeader := packet.VariableLengthHeader.(packets.PublishVariableHeader)
			topic := Topic{
				TopicFilter : varHeader.TopicName,
				Qos: varHeader.,
			handlePublish(server.topicClientMap, topic, clientMessage, server.outputChan)


			// err := handlePublish(server.topicClientMap,)
			// Get the clients connected to that topic and send them
			// all a lovely packet

		case packets.SUBSCRIBE:
			// Then add them to the topic in the subscription table
			fmt.Println("Handling subscribe")
			topics, err := handleSubscribe(server.clientTopicmap, server.topicClientMap, clientID, packet.Payload)
			if err != nil {
				fmt.Println("Received invalid subscribe packet - discarding")
				continue
			}

			returnCodes := make([]byte, len(topics))
			for i, topic := range topics {
				returnCodes[i] = topic.Qos
			}

			packetID := packet.VariableLengthHeader.(packets.SubscribeVariableHeader).PacketIdentifier
			subackArray := packets.CreateSubACK(packetID, returnCodes)
			clientMsg := client.CreateClientMessage(&clientID, &clientConnection, &subackArray)
			(*server.outputChan) <- clientMsg

		case packets.DISCONNECT:
			// Close the client TCP connection.
			// Remove the packet from the client list
			clientConnection.Close()
			delete(clientTable, clientID)
		}
	}

}

// Decode topics and store them in subscription table
func handleSubscribe(clientTopicMap *ClientTopicMap, topicClientMap *TopicClientMap, clientID client.ClientID, packetPayload packets.PacketPayload) ([]Topic, error) {
	topics := make([]Topic, 0, 0)

	payload := packetPayload.ApplicationMessage
	topicNumber, offset := 0, 0

	// FIXME: Make sure we can't send multiple of the same topic in a subscribe message
	// mosquitto_sub -t "test/hello" -t "test/hello" -p 8000

	for offset < len(payload) {
		topicFilter, utfStringLen, err := packets.DecodeUTFString(payload[offset:])

		if err != nil {
			return nil, err
		}

		requestedQOS := payload[offset+utfStringLen]

		topic := Topic{
			TopicFilter: topicFilter,
			Qos:         requestedQOS,
		}
		topics = append(topics, topic)
		topicNumber++
		offset += utfStringLen + 1

		if _, found := (*topicClientMap)[topic]; !found {
			(*topicClientMap)[topic] = make([]client.ClientID, 0, 0)
		}
	}

	if _, found := ((*clientTopicMap)[clientID]); !found {
		(*clientTopicMap)[clientID] = make([]Topic, 0, 0)
	}

	for _, newTopic := range topics {

		clientTopicMap.addClientTopicPair(clientID, newTopic)
		topicClientMap.addTopicClientPair(newTopic, clientID)
	}

	fmt.Println(*clientTopicMap, *topicClientMap)

	return topics, nil

}

func handlePublish(TCMap *TopicClientMap, topic Topic, msgToForward client.ClientMessage, outputChannel *chan client.ClientMessage) {

	clients := (*TCMap)[topic]

	for _, client := range clients {
		msgToForward.ClientID = &client
		(*outputChannel) <- msgToForward
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
