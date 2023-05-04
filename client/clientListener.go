package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// ListenForPackets continually reads packets from the broker connection, decodes them and takes appropriate action.
// For packets that require an ACK, it adds them to the waitingAckStruct.
func (client *Client) ListenForPackets() {
	reader := bufio.NewReader(client.BrokerConnection)

	for {
		packet, err := packets.ReadPacketFromConnection(reader)

		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				structures.PrintCentrally("Connection closed")
				return
			}
			// Quic returns bye message on closing
			if strings.HasSuffix(err.Error(), "bye") {
				return
			}
			if errors.Is(err, io.EOF) {
				return
			}
			fmt.Println("Error during reading:", err)
			return
		}

		packetType := packets.GetPacketType(packet)

		decoded, _, _ := packets.DecodePacket(packet)
		client.ReceivedPackets.Append(decoded)

		switch packetType {
		case packets.SUBACK, packets.CONNACK, packets.PUBACK, packets.UNSUBACK:
			{
				_, offset, _ := packets.DecodeFixedHeader(packet)
				packetID := packets.CombineMsbLsb(packet[offset], packet[offset+1])

				toStore := StoredPacket{
					Packet:   packet,
					PacketID: packetID,
				}
				client.WaitingAckStruct.AddItem(&toStore)
			}

		case packets.PUBLISH:
			{
				result, _, _ := packets.DecodePacket(packet)
				// If we fill the buffer - form a queue
				messageToPrint := result.Payload.RawApplicationMessage[:structures.Min(len(result.Payload.RawApplicationMessage), 20)]
				go fmt.Println("Received request to publish", string(messageToPrint))
			}

		default:
			{
				structures.Println(packet, "Read some packet of type", packets.PacketTypeName(packetType))
			}
		}
	}
}
