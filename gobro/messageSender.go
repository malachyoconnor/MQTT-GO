package gobro

import (
	"MQTT-GO/client"
	"fmt"
)

type MessageSender struct {
	outputChan *chan client.ClientMessage
}

func CreateMessageSender(outputPool *chan client.ClientMessage) MessageSender {
	return MessageSender{
		outputChan: outputPool,
	}
}

func (MessageSender) ListenAndSend(server *Server) {
	for {
		clientMsg := <-(*server.outputChan)

		clientID := *clientMsg.ClientID
		packet := *clientMsg.Packet

		// We look up the client rather than using the connection directly
		// This is to ensure we get an error if the client doesn't exist

		client := (*server.clientTable)[clientID]
		_, err := client.TCPConnection.Write(packet)

		if err != nil {
			fmt.Println("Failed to send packet. ", err)
		}

		continue

	}

}
