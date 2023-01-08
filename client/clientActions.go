package client

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"fmt"
	"net"
)

// Here we'll have the functions that make the client perform it's actions.

func ConnectToServer(ip string, port int) error {
	address := net.JoinHostPort(ip, fmt.Sprint(port))
	connection, err := net.Dial("tcp", address)

	if err != nil {
		return err
	}

	controlHeader := packets.ControlHeader{Type: packets.CONNECT, Flags: 0}
	varHeader := packets.ConnectVariableHeader{}
	payload := packets.PacketPayload{}
	payload.ClientID = "testing"
	connectPacket := packets.CombinePacketSections(&controlHeader, varHeader, &payload)
	connectPacketArr, err := packets.CreateConnect(connectPacket)

	if err != nil {
		return err
	}
	connection.Write(*connectPacketArr)
	reader := bufio.NewReader(connection)
	result, _ := packets.ReadPacketFromConnection(reader)
	packet, _, _ := packets.DecodePacket(*result)
	structures.PrintInterface(*packet)

	return nil
}
