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
		case packets.SUBACK, packets.CONNACK, packets.PUBACK:
			{
				_, offset, _ := packets.DecodeFixedHeader(packet)
				packetID := packets.CombineMsbLsb(packet[offset], packet[offset+1])

				toStore := storedPacket{
					packet:   packet,
					packetID: packetID,
				}
				client.waitingPackets.AddItem(&toStore)
				fmt.Println("Stored in", packetID)

			}
		default:
			{
				fmt.Println(packet, "Read some random packet of type", packets.PacketTypeName(packetType))
			}
		}

	}

}
