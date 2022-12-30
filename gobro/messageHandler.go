package gobro

import (
	"MQTT-GO/clients"
	"MQTT-GO/packets"
	"fmt"
)

type MessageHandler struct {
	AttachedInputChan  *chan clients.ClientMessage
	AttachedOutputChan *chan clients.ClientMessage
}

func CreateMessageHandler(inputChanAddress *chan clients.ClientMessage, outputChanAddress *chan clients.ClientMessage) MessageHandler {
	return MessageHandler{
		AttachedInputChan:  inputChanAddress,
		AttachedOutputChan: outputChanAddress,
	}
}

func (msgH *MessageHandler) Listen(server *Server) {

	clientTable := server.clientTable

	for {
		clientMessage := <-(*msgH.AttachedInputChan)
		clientID := *clientMessage.ClientID
		// Could be nil if the client doesn't exist yet
		client := (*clientTable)[clientID]

		packetArray := clientMessage.Packet
		packet, packetType, err := packets.DecodePacket(*packetArray)

		if err != nil {
			fmt.Println(err)
			continue
		}
		// General case for if the client doesn't exist IF NOT A CONNECT packet
		if packetType != packets.CONNECT {

			if _, found := (*clientTable)[clientID]; !found {
				fmt.Println("Client not in the client table sent", packets.PacketTypeName(packetType), "message, disconnecting.")
				// If the client hasn't already been disconnected by the client handler
				if client != nil {
					client.TCPConnection.Close()
				}
				continue
			}

		}

		go HandleMessage(packetType, packet, client, server, clientMessage)

	}

}

func HandleMessage(packetType byte, packet *packets.Packet, client *clients.Client, server *Server, clientMessage clients.ClientMessage) {
	clientTable := server.clientTable
	clientID := *clientMessage.ClientID
	clientConnection := *clientMessage.ClientConnection
	if client != nil {
		client.Queue.DoingWork()
	}

	switch packetType {
	case packets.CONNECT:
		// Query the client table to check if the client exists
		// if not slap it in there - then send connack

		// This should disconnect them if they're already connected !!
		createClient(clientTable, &clientMessage)
		// Check if the reserved flag is zero, if not disconnect them
		// Finally send out a CONACK [X]
		connackArray := packets.CreateConnACK(false, 0)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, &connackArray)
		(*server.outputChan) <- clientMsg

	case packets.PUBLISH:
		fmt.Println("Received request to publish:", string(packet.Payload.ApplicationMessage[:]))

		varHeader := packet.VariableLengthHeader.(packets.PublishVariableHeader)
		topic := clients.Topic{
			TopicFilter: varHeader.TopicFilter,
			Qos:         packet.ControlHeader.Flags & 6,
		}
		handlePublish(server.topicClientMap, topic, clientMessage, server.outputChan, server.clientTable)

		// err := handlePublish(server.topicClientMap,)
		// Get the clients connected to that topic and send them
		// all a lovely packet

	case packets.SUBSCRIBE:
		fmt.Println("Handling subscribe")
		// Add the client to the topic in the subscription table
		topics, err := handleSubscribe(server.topicClientMap, client, packet.Payload)
		if err != nil {
			fmt.Println("Error during subscribe:", err)
			return
		}

		// Get the return code for every topic
		returnCodes := make([]byte, len(topics))
		for i, topic := range topics {
			returnCodes[i] = topic.Qos
		}

		packetID := packet.VariableLengthHeader.(packets.SubscribeVariableHeader).PacketIdentifier
		subackPacket := packets.CreateSubACK(packetID, returnCodes)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, &subackPacket)
		(*server.outputChan) <- clientMsg

	case packets.DISCONNECT:
		// Close the client TCP connection.
		// Remove the packet from the client list
		fmt.Println("Disconnecting", clientID)
		clientConnection.Close()

		// TODO - do this centrally.

		delete(*clientTable, clientID)
	}

	if (packetType != packets.PUBLISH) && (client != nil) {
		client.Queue.FinishedWork()
	}
}

// Decode topics and store them in subscription table
func handleSubscribe(topicClientMap *clients.TopicClientMap, client *clients.Client, packetPayload packets.PacketPayload) ([]clients.Topic, error) {
	newTopics := make([]clients.Topic, 0, 0)
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

		topic := clients.Topic{
			TopicFilter: topicFilter,
			Qos:         requestedQOS,
		}
		newTopics = append(newTopics, topic)
		topicNumber++
		offset += utfStringLen + 1

		if _, found := (*topicClientMap)[topic]; !found {
			continue // remove
			// (*topicClientMap)[topic] = make([]clients.ClientID, 0, 0)
		}
	}

	clientTopics := client.Topics

	if clientTopics == nil {
		clientTopics = make([]clients.Topic, 0, 0)
	}

	for _, newTopic := range newTopics {

		client.AddTopic(newTopic)
		topicClientMap.AddTopicClientPair(newTopic, client.ClientIdentifier)
	}

	fmt.Println(client.Topics, *topicClientMap)

	return newTopics, nil

}

func handlePublish(TCMap *clients.TopicClientMap, topic clients.Topic, msgToForward clients.ClientMessage, outputChannel *chan clients.ClientMessage, clientTable *clients.ClientTable) {
	return
	// clientList := (*TCMap)[topic]

	// for i := range clientList {

	// 	clientID := clientList[i]
	// 	alteredMsg := msgToForward
	// 	alteredMsg.ClientID = &clientID
	// 	alteredMsg.ClientConnection = &(*clientTable)[clientID].TCPConnection

	// 	(*outputChannel) <- alteredMsg
	// }

}

func createClient(clientTable *clients.ClientTable, clientMsg *clients.ClientMessage) *clients.Client {

	if client, found := (*clientTable)[*clientMsg.ClientID]; found {
		return client
	}

	client := clients.CreateClient(*clientMsg.ClientID, clientMsg.ClientConnection)
	(*clientTable)[*clientMsg.ClientID] = &client

	return &client
}
