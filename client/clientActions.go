package client

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"fmt"
)

// Here we'll have the functions that make the client perform it's actions.

func (client *Client) SendConnect() error {

	controlHeader := packets.ControlHeader{Type: packets.CONNECT, Flags: 0}
	varHeader := packets.ConnectVariableHeader{}
	payload := packets.PacketPayload{}
	payload.ClientID = "testing"
	connectPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	connectPacketArr, err := packets.CreateConnect(connectPacket)

	if err != nil {
		return err
	}
	(*client.brokerConnection).Write(*connectPacketArr)
	reader := bufio.NewReader(*client.brokerConnection)
	result, _ := packets.ReadPacketFromConnection(reader)
	packet, _, _ := packets.DecodePacket(*result)
	structures.PrintInterface(*packet)

	return nil
}

// TODO: Handle readPacketFromConnection error properly
// TODO: Check if everything you would need for a publish packet is present!

func (client *Client) SendPublish(applicationMessage *[]byte, topic string) error {

	controlHeader := packets.ControlHeader{Type: packets.PUBLISH, Flags: 0}
	varHeader := packets.PublishVariableHeader{}
	varHeader.TopicFilter = topic
	payload := packets.PacketPayload{}
	payload.ApplicationMessage = *applicationMessage

	publishPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	publishPacketArr, err := packets.CreatePublish(publishPacket)

	if err != nil {
		return err
	}

	(*client.brokerConnection).Write(*publishPacketArr)

	// Check the qos level to see if we should expect a response - if not then exit
	if controlHeader.Flags&6 == 0 {
		fmt.Println(controlHeader.Flags)
		return nil
	}

	reader := bufio.NewReader(*client.brokerConnection)
	result, err := packets.ReadPacketFromConnection(reader)

	if err != nil {
		return err
	}

	packet, _, _ := packets.DecodePacket(*result)
	structures.PrintInterface(*packet)

	return nil
}
