package gobro

import (
	"MQTT-GO/gobro/clients"
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
	for {
		clientMsg := <-(*server.outputChan)
		clientID := *clientMsg.ClientID

		client := server.clientTable.Get(clientID)
		if client == nil {
			continue
		}

		ticket := client.Tickets.GetTicket()
		packet := clientMsg.Packet

		// We look up the client rather than using the connection directly
		// This is to ensure we get an error if the client doesn't exist

		go func() {
			ticket.WaitOnTicket()
			_, err := (*clientMsg.ClientConnection).Write(packet)
			ticket.TicketCompleted()
			if err != nil {
				clients.ServerPrintln("Failed to send packet to", client, "- Error:", err)
			}
		}()

		continue
	}
}
