package gobro

import (
	"MQTT-GO/gobro/clients"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
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
		client := clientTable.Get(clientID)

		if client == nil {
			fmt.Printf("Client '%v', who no longer exists, sent a message", clientID)
			continue
		}

		ticket := client.Tickets.GetTicket()
		packetArray := clientMessage.Packet
		packet, packetType, err := packets.DecodePacket(packetArray)

		if err != nil {
			fmt.Println(err)
			continue
		}
		// General case for if the client doesn't exist if NOT a connect packet
		if packetType != packets.CONNECT {

			if !clientTable.Contains(clientID) {
				fmt.Println("Client not in the client table sent", packets.PacketTypeName(packetType), "message, disconnecting.")
				// If the client hasn't already been disconnected by the client handler
				if client != nil {
					client.TCPConnection.Close()
				}
				continue
			}

		}

		go HandleMessage(packetType, packet, client, server, clientMessage, ticket)

	}

}

func HandleMessage(packetType byte, packet *packets.Packet, client *clients.Client, server *Server, clientMessage clients.ClientMessage, ticket structures.Ticket) {
	clientTable := server.clientTable
	clientID := *clientMessage.ClientID
	clientConnection := *clientMessage.ClientConnection
	packetsToSend := make([]*clients.ClientMessage, 0, 10)

	switch packetType {
	case packets.CONNECT:
		// Check if the reserved flag is zero, if not disconnect them
		// Finally send out a CONACK [X]
		connack := packets.CreateConnACK(false, 0)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, connack)
		packetsToSend = append(packetsToSend, &clientMsg)

	case packets.PUBLISH:
		fmt.Println("Received request to publish:", string(packet.Payload.ApplicationMessage[:]), "to topic:", string(packet.VariableLengthHeader.(*packets.PublishVariableHeader).TopicFilter))

		varHeader := packet.VariableLengthHeader.(*packets.PublishVariableHeader)
		topic := clients.Topic{
			TopicFilter: varHeader.TopicFilter,
			Qos:         packet.ControlHeader.Flags & 6,
		}
		// Adds to the packets to send
		handlePublish(server.topicClientMap, topic, clientMessage, server.outputChan, server.clientTable, &packetsToSend)

	case packets.SUBSCRIBE:
		// Add the client to the topic in the subscription table
		topics, err := handleSubscribe(server.topicClientMap, client, *packet.Payload)
		if err != nil {
			fmt.Println("Error during subscribe:", err)
			return
		}

		// Get the return code for every topic
		returnCodes := make([]byte, len(topics))
		for i, topic := range topics {
			returnCodes[i] = topic.Qos
		}

		packetID := packet.VariableLengthHeader.(*packets.SubscribeVariableHeader).PacketIdentifier
		subackPacket := packets.CreateSubACK(packetID, returnCodes)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, subackPacket)
		packetsToSend = append(packetsToSend, &clientMsg)

	case packets.DISCONNECT:
		// Close the client TCP connection.
		// Remove the packet from the client list
		ticket.WaitOnTicket()
		client.Disconnect(server.topicClientMap, clientTable)
		ticket.TicketCompleted()
		return
	}

	ticket.WaitOnTicket()
	for _, packet := range packetsToSend {

		(*server.outputChan) <- *packet

	}
	ticket.TicketCompleted()
}

// Decode topics and store them in subscription table
func handleSubscribe(topicClientMap *clients.TopicToSubscribers, client *clients.Client, packetPayload packets.PacketPayload) ([]clients.Topic, error) {
	newTopics := make([]clients.Topic, 0)
	payload := packetPayload.ApplicationMessage
	topicNumber, offset := 0, 0

	// Progress through the payload and read every topic & QoS level that the client wants to subscribe to
	// Then add them to a list to be handled
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

		if !topicClientMap.Contains(topic.TopicFilter) {
			topicClientMap.AddTopic(topic.TopicFilter)
		}
	}

	clientTopics := client.Topics

	if clientTopics == nil {
		client.Topics = structures.CreateLinkedList[clients.Topic]()
	}

	for _, newTopic := range newTopics {
		client.AddTopic(newTopic)
		topicClientMap.Put(newTopic.TopicFilter, client.ClientIdentifier)
		structures.PrintCentrally("SUBSCRIBED TO ", newTopic.TopicFilter)
	}

	return newTopics, nil

}

func handlePublish(TCMap *clients.TopicToSubscribers, topic clients.Topic, msgToForward clients.ClientMessage, outputChannel *chan clients.ClientMessage, clientTable *structures.SafeMap[clients.ClientID, *clients.Client], toSend *[]*clients.ClientMessage) {
	clientList, err := TCMap.GetMatchingClients(topic.TopicFilter)

	if err != nil {
		fmt.Println(err)
		return
	}
	clientNode := clientList.Head()

	for clientNode != nil {
		clientID := clientNode.Value()
		alteredMsg := msgToForward
		alteredMsg.ClientID = &clientID

		if client := clientTable.Get(clientID); client != nil {
			alteredMsg.ClientConnection = &(client.TCPConnection)
		} else {
			fmt.Printf("error: Can't find subscribed client '%v' in clientTable", clientID)
			clientNode = clientNode.Next()
			continue
		}

		(*toSend) = append(*toSend, &alteredMsg)
		clientNode = clientNode.Next()
	}

}
