package gobro

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"fmt"
)

type MessageSender struct {
	outputChan *chan []byte
}

func CreateMessageSender(outputPool *chan []byte) MessageSender {
	return MessageSender{
		outputChan: outputPool,
	}
}

func (MessageSender) ListenAndSend(server *Server) {
	for {
		packetToSend := <-(*server.outputChan)

		fmt.Println("Sending some shit")

		stringID, idLen, err := packets.DecodeUTFString(packetToSend)
		clientID := client.ClientID(stringID)
		if err != nil {
			fmt.Println("Failed to decode username. ", err)
			continue
		}

		client := (*server.clientTable)[clientID]
		_, err = client.TCPConnection.Write(packetToSend[idLen:])

		if err != nil {
			fmt.Println("Failed to send packet. ", err)
		}
		continue

	}

}
