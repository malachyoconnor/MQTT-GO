package client

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"errors"
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
	if client.brokerConnection == nil {
		return errors.New("error: Client does not have a broker connection")
	}

	_, err = (*client.brokerConnection).Write(*connectPacketArr)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(*client.brokerConnection)
	result, _ := packets.ReadPacketFromConnection(reader)
	packet, _, _ := packets.DecodePacket(*result)

	if packet.ControlHeader.Type != packets.CONNACK {
		structures.PrintInterface(packet)
		return errors.New("error: Received packet other than CONNACK from server")
	} else {
		if packet.VariableLengthHeader.(*packets.ConnackVariableHeader).ConnectReturnCode == 2 {
			return errors.New("error: Server already contains this clientID")
		}
	}

	return nil
}

// TODO: Handle readPacketFromConnection error properly
// TODO: Check if everything you would need for a publish packet is present!

func (client *Client) SendPublish(applicationMessage []byte, topic string) error {

	controlHeader := packets.ControlHeader{Type: packets.PUBLISH, Flags: 0}
	varHeader := packets.PublishVariableHeader{}
	varHeader.TopicFilter = topic
	payload := packets.PacketPayload{}
	payload.ApplicationMessage = applicationMessage

	publishPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	publishPacketArr, err := packets.CreatePublish(publishPacket)

	if err != nil {
		return err
	}
	// TODO: Get the err from this
	_, err = (*client.brokerConnection).Write(*publishPacketArr)

	if err != nil {
		return err
	}

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

	if packet.ControlHeader.Type != packets.PUBACK {
		return errors.New("error: Didn't receive PUBACK from server")
	}

	return nil
}

func (client *Client) SendDisconnect() error {
	controlHeader := packets.ControlHeader{}
	controlHeader.Flags = 0
	controlHeader.Type = packets.DISCONNECT
	controlHeader.RemainingLength = 0

	disconnectArr := packets.EncodeFixedHeader(controlHeader)
	_, err := (*client.brokerConnection).Write(disconnectArr)

	if err != nil {
		return err
	}
	return nil
}
