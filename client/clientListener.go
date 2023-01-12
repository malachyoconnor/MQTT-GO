package client

import (
	"bufio"
	"fmt"

	"MQTT-GO/packets"
)

func (client *Client) ListenForPackets() {
	reader := bufio.NewReader(*client.BrokerConnection)
	for {
		packet, err := packets.ReadPacketFromConnection(reader)
		if err != nil {
			fmt.Println(err)
		}

		packetType := packets.GetPacketType(packet)
		fmt.Println("Received", packets.PacketTypeName(packetType))

		switch packetType {
		case packets.SUBACK, packets.CONNACK, packets.PUBACK, packets.UNSUBACK:
			{
				_, offset, _ := packets.DecodeFixedHeader(packet)
				packetID := packets.CombineMsbLsb(packet[offset], packet[offset+1])

				toStore := StoredPacket{
					Packet:   packet,
					PacketID: packetID,
				}
				client.waitingPackets.AddItem(&toStore)
			}

		case packets.PUBLISH:
			{
				result, _, _ := packets.DecodePacket(packet)
				fmt.Println("Received request to publish", string(result.Payload.RawApplicationMessage))
			}

		default:
			{
				fmt.Println(packet, "Read some packet of type", packets.PacketTypeName(packetType))
			}
		}

	}
}
