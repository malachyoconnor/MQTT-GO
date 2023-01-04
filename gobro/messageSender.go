package gobro

import (
	"MQTT-GO/clients"
	"fmt"
)

type MessageSender struct {
	outputChan *chan clients.ClientMessage
}

func CreateMessageSender(outputPool *chan clients.ClientMessage) MessageSender {
	return MessageSender{
		outputChan: outputPool,
	}
}

func (MessageSender) ListenAndSend(server *Server) {
	for {
		clientMsg := <-(*server.outputChan)
		clientID := *clientMsg.ClientID

		client := server.clientTable.Get(clientID)
		if client == nil {
			continue
		}

		ticket := client.Tickets.GetTicket()
		packet := *clientMsg.Packet

		// packetType := packets.PacketTypeName(packets.GetPacketType(&packet))
		// fmt.Println("Sending", packetType, "to", clientID)
		// We look up the client rather than using the connection directly
		// This is to ensure we get an error if the client doesn't exist

		go func() {
			ticket.WaitOnTicket()
			_, err := (*clientMsg.ClientConnection).Write(packet)
			ticket.TicketCompleted()
			if err != nil {
				fmt.Println("Failed to send packet to", client, "- Error:", err)
			}
		}()

		continue

	}

}
