package client

import (
	"bufio"
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
				// Quic returns bye message on closing
			}
			if strings.HasSuffix(err.Error(), "bye") {
				return
			}
			structures.Println("Error during reading:", err)
			return
		}

		packetType := packets.GetPacketType(packet)
		structures.Println("Received", packets.PacketTypeName(packetType))

		switch packetType {
		case packets.SUBACK, packets.CONNACK, packets.PUBACK, packets.UNSUBACK:
			{
				_, offset, _ := packets.DecodeFixedHeader(packet)
				packetID := packets.CombineMsbLsb(packet[offset], packet[offset+1])

				toStore := StoredPacket{
					Packet:   packet,
					PacketID: packetID,
				}
				client.waitingAckStruct.AddItem(&toStore)
			}

		case packets.PUBLISH:
			{
				structures.Println("READ A PUBLISH")
				result, _, _ := packets.DecodePacket(packet)
				// If we fill the buffer - form a queue
				if len(client.ReceivedPackets) == waitingPacketBufferSize {
					go func() { client.ReceivedPackets <- result }()
				}
				client.ReceivedPackets <- result
				go structures.Println("Received request to publish", string(result.Payload.RawApplicationMessage))
			}

		default:
			{
				structures.Println(packet, "Read some packet of type", packets.PacketTypeName(packetType))
			}
		}
	}
}
