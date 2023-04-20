package gobro

import (
	"log"

	"MQTT-GO/gobro/clients"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// MessageHandler is a struct that handles the messages that are sent to the server.
// It has a channel for incoming packets, and a channel for outgoing packets.
type MessageHandler struct {
	AttachedInputChan  *chan clients.ClientMessage
	AttachedOutputChan *chan clients.ClientMessage
}

// CreateMessageHandler creates a new message handler with a channel for incoming packets,
// and a channel for outgoing packets.
func CreateMessageHandler(inputChan *chan clients.ClientMessage,
	outputChan *chan clients.ClientMessage) MessageHandler {
	return MessageHandler{
		AttachedInputChan:  inputChan,
		AttachedOutputChan: outputChan,
	}
}

// Listen listens for incoming packets, decodes them and then runs the HandleMessage function
// in a separate goroutine.
func (msgH *MessageHandler) Listen(server *Server) {
	clientTable := server.clientTable

	for {
		clientMessage := <-(*msgH.AttachedInputChan)
		clientID := *clientMessage.ClientID
		client := clientTable.Get(clientID)

		if client == nil {
			packetType := packets.PacketTypeName(packets.GetPacketType(clientMessage.Packet))
			log.Printf("- Client '%v', who no longer exists, sent a %v packet\n", clientID, packetType)
			continue
		}

		ticket := client.Tickets.GetTicket()
		packetArray := clientMessage.Packet
		packet, packetType, err := packets.DecodePacket(packetArray)
		if err != nil {
			log.Printf("- Error during decoding '%v', from '%v': %v\n", packets.PacketTypeName(packetType), clientID, err)
			continue
		}
		// General case for if the client doesn't exist if NOT a connect packet
		if packetType != packets.CONNECT {
			if !clientTable.Contains(clientID) {
				log.Printf("Client '%v' not in the client table sent %v message, disconnecting.\n",
					clientID, packets.PacketTypeName(packetType))
				structures.Printf("Client '%v' not in the client table sent %v message, disconnecting.\n",
					clientID, packets.PacketTypeName(packetType))

				// If the client hasn't already been disconnected by the client handler
				client.NetworkConnection.Close()
				continue
			}
		}
		go HandleMessage(packetType, packet, client, server, clientMessage, ticket)
	}
}

// HandleMessage handles the incoming packets by checking the packet type and then
// running the appropriate function.
// It also encodes the outgoing packets and sends them to the MessageSender.
func HandleMessage(packetType byte, packet *packets.Packet, client *clients.Client, server *Server,
	clientMessage clients.ClientMessage, ticket structures.Ticket) {
	clientTable := server.clientTable
	clientID := *clientMessage.ClientID
	clientConnection := *clientMessage.ClientConnection
	packetsToSend := make([]*clients.ClientMessage, 0, 10)
	topicClientMap := server.topicClientMap

	switch packetType {
	case packets.CONNECT:
		// Check if the reserved flag is zero, if not disconnect them
		// Finally send out a CONACK [X]
		connack := packets.CreateConnACK(false, 0)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, connack)
		packetsToSend = append(packetsToSend, &clientMsg)

	case packets.PUBLISH:
		varHeader, ok := packet.VariableLengthHeader.(*packets.PublishVariableHeader)
		if !ok {
			log.Printf("Error during publish, from client '%v'\n", clientID)
			return
		}
		topic := clients.Topic{
			TopicFilter: varHeader.TopicFilter,
			Qos:         packet.ControlHeader.Flags & 6,
		}
		msgToPublish := string(packet.Payload.RawApplicationMessage)
		go structures.Println("Received request to publish:", msgToPublish, "to topic:", topic.TopicFilter)

		// Adds to the packets to send
		handlePublish(topicClientMap, topic, clientMessage, server.clientTable, &packetsToSend)

	case packets.SUBSCRIBE:
		// Add the client to the topic in the subscription table
		topics, err := handleSubscribe(topicClientMap, client, *packet.Payload)
		if err != nil {
			log.Printf("Error during subscribe: %v, from client '%v'\n", err, clientID)
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

	case packets.UNSUBSCRIBE:
		packetID := packet.VariableLengthHeader.(*packets.UnsubscribeVariableHeader).PacketIdentifier
		// Note that in an unsubscribe we don't need the QOS levels
		topics := make([]string, 0, len(packet.Payload.TopicList))
		for _, topic := range packet.Payload.TopicList {
			topics = append(topics, topic.Topic)
		}
		handleUnsubscribe(topics, topicClientMap, *client)
		unsubackPacket := packets.CreateUnSuback(packetID)
		clientMsg := clients.CreateClientMessage(clientID, &clientConnection, unsubackPacket)
		packetsToSend = append(packetsToSend, &clientMsg)

	case packets.DISCONNECT:
		// Close the client connection.
		// Remove the packet from the client list
		ticket.WaitOnTicket()
		go client.Disconnect(topicClientMap, clientTable)
		ticket.TicketCompleted()
		return
	}

	ticket.WaitOnTicket()
	for _, packet := range packetsToSend {
		(*server.outputChan) <- *packet
	}
	ticket.TicketCompleted()
}

// Decode topics and store them in subscription table.
func handleSubscribe(topicClientMap *clients.TopicToSubscribers,
	client *clients.Client, packetPayload packets.PacketPayload) ([]clients.Topic, error) {
	newTopics := make([]clients.Topic, 0)
	payload := packetPayload.RawApplicationMessage
	topicNumber, offset := 0, 0

	// Progress through the payload and read every topic & QoS level that the client wants to subscribe to
	// Then add them to a list to be handled.
	for offset < len(payload) {
		topicFilter, utfStringLen, err := packets.DecodeUTFString(payload[offset:])
		if err != nil {
			structures.Println("Error decoding UTF string")
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
			err = topicClientMap.AddTopic(topic.TopicFilter)
			if err != nil {
				log.Printf("- Error while adding new topic %v, the topic name was '%v'\n", err, topicFilter)
				return nil, err
			}
		}
	}
	if client.Topics == nil {
		client.Topics = structures.CreateLinkedList[clients.Topic]()
	}

	topicClientMap.PrintTopics()

	for _, newTopic := range newTopics {
		client.AddTopic(newTopic)
		err := topicClientMap.Put(newTopic.TopicFilter, client.ClientIdentifier)
		if err != nil {
			log.Printf("- Error while adding new topic %v, the topic name was '%v'\n", err, newTopic.TopicFilter)
			return nil, err
		}
		structures.PrintCentrally("SUBSCRIBED TO ", newTopic.TopicFilter)
	}

	return newTopics, nil
}

func handleUnsubscribe(topics []string, topicToSubscribers *clients.TopicToSubscribers, client clients.Client) {
	topicToSubscribers.Unsubscribe(client.ClientIdentifier, topics...)
	for _, topic := range topics {
		err := client.RemoveTopic(clients.Topic{TopicFilter: topic})
		log.Printf("- Error while removing '%v' from client topic list: %v \n", client.ClientIdentifier, err)
	}
}

func handlePublish(tCMap *clients.TopicToSubscribers, topic clients.Topic, msgToForward clients.ClientMessage,
	clientTable *structures.SafeMap[clients.ClientID, *clients.Client], toSend *[]*clients.ClientMessage) {
	clientList, err := tCMap.GetMatchingClients(topic.TopicFilter)

	if err != nil {
		log.Printf("- Error while getting matching clients during a publish to '%v' by '%v': %v\n",
			topic.TopicFilter, *msgToForward.ClientID, err)
		return
	}
	clientNode := clientList.Head()

	for clientNode != nil {
		clientID := clientNode.Value()
		alteredMsg := msgToForward
		alteredMsg.ClientID = &clientID

		if client := clientTable.Get(clientID); client != nil {
			alteredMsg.ClientConnection = &(client.NetworkConnection)
		} else {
			log.Printf("- Error: Can't find subscribed client '%v' in clientTable\n", clientID)
			clientNode = clientNode.Next()
			continue
		}

		(*toSend) = append(*toSend, &alteredMsg)
		clientNode = clientNode.Next()
	}
}
