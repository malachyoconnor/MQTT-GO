package gobro

import (
	"MQTT-GO/gobro/clients"
	"fmt"
	"sync"
)

// MessageSender is a struct that handles outgoing packets from the broker.
type MessageSender struct {
	outputChan *chan clients.ClientMessage
}

// CreateMessageSender creates a new message sender with a channel for outgoing packets.
func CreateMessageSender(outputPool *chan clients.ClientMessage) MessageSender {
	return MessageSender{
		outputChan: outputPool,
	}
}

// ListenAndSend listens for outgoing packets, finds the appropriate client and sends them.
// It waits for a ticket to be available before sending the packet to ensure messages.
// are sent in the correct order
func (MessageSender) ListenAndSend(server *Server) {
	const maxMessageSenders = 8000
	queue := make(chan struct{}, maxMessageSenders)

	for i := 0; i < maxMessageSenders; i++ {
		queue <- struct{}{}
	}

	for {
		clientMsg := <-(*server.outputChan)

		clientID := *clientMsg.ClientID
		client := server.clientTable.Get(clientID)
		if client == nil {
			fmt.Println("Nil client")
			continue
		}
		// Wait for there to be space in the queue (when a thread has finished and added to it)
		<-queue
		go waitAndSend(&clientMsg, clientMsg.OutputWaitGroup, queue)
		// We look up the client rather than using the connection directly
		// This is to ensure we get an error if the client doesn't exist
	}
}

func waitAndSend(clientMsg *clients.ClientMessage, waitGroup *sync.WaitGroup, queue chan struct{}) {

	defer waitGroup.Done()
	_, err := clientMsg.ClientConnection.Write(clientMsg.Packet)

	if err != nil {
		clients.ServerPrintln("Failed to send packet to", clientMsg.ClientID, "- Error:", err)
	}
	queue <- struct{}{}
}
