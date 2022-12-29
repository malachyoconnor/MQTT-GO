package gobro

import (
	"MQTT-GO/clients"
	"MQTT-GO/packets"
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
		packet := *clientMsg.Packet

		packetType := packets.PacketTypeName(packets.GetPacketType(&packet))
		fmt.Println("Sending", packetType, "to", clientID)
		// We look up the client rather than using the connection directly
		// This is to ensure we get an error if the client doesn't exist

		client, found := (*server.clientTable)[clientID]
		if !found {
			continue
		}

		go func() {
			client.Queue.JoinWaitList()
			defer client.Queue.FinishedWork()
			_, err := client.TCPConnection.Write(packet)

			if err != nil {
				fmt.Println("Failed to send packet to", client, "- Error:", err)
			}
		}()

		continue

	}

}
