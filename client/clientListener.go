package client

import (
	"MQTT-GO/packets"
	"bufio"
	"fmt"
)

func (client *Client) ListenForPackets() {
	reader := bufio.NewReader(*client.brokerConnection)
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

				toStore := storedPacket{
					packet:   packet,
					packetID: packetID,
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
