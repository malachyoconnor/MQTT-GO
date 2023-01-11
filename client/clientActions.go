package client

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"errors"
	"fmt"
	"time"
)

// Here we'll have the functions that make the client perform it's actions.

func (client *Client) SendConnect() error {

	if client.brokerConnection == nil {
		return errors.New("error: Client does not have a broker connection")
	}

	controlHeader := packets.ControlHeader{Type: packets.CONNECT, Flags: 0}
	varHeader := packets.ConnectVariableHeader{}
	payload := packets.PacketPayload{}
	payload.ClientID = client.clientID

	connectPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	connectPacketArr, err := packets.EncodeConnect(connectPacket)
	if err != nil {
		return err
	}

	_, err = (*client.brokerConnection).Write(connectPacketArr)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(*client.brokerConnection)
	result, err := packets.ReadPacketFromConnection(reader)
	fmt.Println(result, "Read connack")
	if err != nil {
		return err
	}
	packet, _, _ := packets.DecodePacket(result)

	if packet.ControlHeader.Type != packets.CONNACK {
		structures.PrintInterface(packet)
		return errors.New("error: Received packet other than CONNACK from server")
	} else {
		// If the clientID already exists then we wait
		if packet.VariableLengthHeader.(*packets.ConnackVariableHeader).ConnectReturnCode == 2 {
			time.Sleep(time.Millisecond)
			client.clientID = generateRandomClientID()
			client.SetClientConnection(*ip, *port)
			return client.SendConnect()
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
	varHeader.PacketIdentifier = getAndIncrementPacketId()
	payload := packets.PacketPayload{}
	payload.ApplicationMessage = applicationMessage

	publishPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	publishPacketArr, err := packets.EncodePublish(publishPacket)

	if err != nil {
		return err
	}
	_, err = (*client.brokerConnection).Write(publishPacketArr)

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

	packet, _, _ := packets.DecodePacket(result)

	if packet.ControlHeader.Type != packets.PUBACK {
		return errors.New("error: Didn't receive PUBACK from server")
	}

	return nil
}

type topicWithQoS struct {
	topic string
	QoS   byte
}

func (client *Client) SendSubscribe(topics ...topicWithQoS) error {

	controlHeader := packets.ControlHeader{Type: packets.SUBSCRIBE, Flags: 2}
	varHeader := packets.SubscribeVariableHeader{}
	packetID := getAndIncrementPacketId()
	varHeader.PacketIdentifier = packetID
	payload := packets.PacketPayload{}
	payload.ApplicationMessage = make([]byte, 0, 2*len(topics))

	for _, topicWQos := range topics {
		if topicWQos.QoS > 2 {
			return errors.New("error: impossible QoS level provided")
		}
		encodedTopic, _, err := packets.EncodeUTFString(topicWQos.topic)
		if err != nil {
			return err
		}
		payload.ApplicationMessage = append(payload.ApplicationMessage, encodedTopic...)
		payload.ApplicationMessage = append(payload.ApplicationMessage, topicWQos.QoS)
	}

	packet := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	encodedPacket, err := packets.EncodeSubscribe(packet)

	if err != nil {
		return err
	}
	(*client.brokerConnection).Write(encodedPacket)

	// Working - but I need to come up with some functions to listen for our
	// packet identifier innit!

	subackArr := client.waitingPackets.GetOrWait(packetID)
	suback, _, _ := packets.DecodePacket(*subackArr)

	if suback.ControlHeader.Type != packets.SUBACK {
		return errors.New("error: Our SUBACK got nabbed")
	}

	returnCodes := (*subackArr)[4:]
	fmt.Println(returnCodes)

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
