package client

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// Here we'll have the functions that make the client perform it's actions.

func (client *Client) SendConnect() error {
	if client.BrokerConnection == nil {
		return errors.New("error: Client does not have a broker connection")
	}

	controlHeader := packets.ControlHeader{Type: packets.CONNECT, Flags: 0}
	varHeader := packets.ConnectVariableHeader{}
	payload := packets.PacketPayload{}
	payload.ClientID = client.ClientID

	connectPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	connectPacketArr, err := packets.EncodeConnect(connectPacket)
	if err != nil {
		return err
	}
	if client.BrokerConnection == nil {
		return errors.New("error: connection is closed")
	}
	_, err = (*client.BrokerConnection).Write(connectPacketArr)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(*client.BrokerConnection)
	result, err := packets.ReadPacketFromConnection(reader)
	fmt.Println(result, "Read connack")
	if err != nil {
		return err
	}
	packet, _, _ := packets.DecodePacket(result)

	if packet.ControlHeader.Type != packets.CONNACK {
		structures.PrintInterface(packet)
		return errors.New("error: Received packet other than CONNACK from server")
	} else if packet.VariableLengthHeader.(*packets.ConnackVariableHeader).ConnectReturnCode == 2 {
		// If the clientID already exists then we wait
		time.Sleep(time.Millisecond)
		client.ClientID = generateRandomClientID()
		err := client.SetClientConnection(*ip, *port)
		if err != nil {
			fmt.Println("Error while setting client connection")
			return err
		}
		return client.SendConnect()
	}

	return nil
}

// TODO: Handle readPacketFromConnection error properly
// TODO: Check if everything you would need for a publish packet is present!

func (client *Client) SendPublish(applicationMessage []byte, topic string) error {
	controlHeader := packets.ControlHeader{Type: packets.PUBLISH, Flags: 0}
	varHeader := packets.PublishVariableHeader{}
	varHeader.TopicFilter = topic
	packetID := getAndIncrementPacketID()
	varHeader.PacketIdentifier = packetID
	payload := packets.PacketPayload{}
	payload.RawApplicationMessage = applicationMessage

	publishPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	publishPacketArr, err := packets.EncodePublish(publishPacket)
	if err != nil {
		return err
	}
	if client.BrokerConnection == nil {
		return errors.New("error: connection is closed")
	}
	_, err = (*client.BrokerConnection).Write(publishPacketArr)

	if err != nil {
		return err
	}

	// Check the qos level to see if we should expect a response - if not then exit
	if controlHeader.Flags&6 == 0 {
		fmt.Println(controlHeader.Flags)
		return nil
	}

	pubackArr := client.waitingPackets.GetOrWait(packetID)
	packet, _, _ := packets.DecodePacket(*pubackArr)

	if packet.ControlHeader.Type != packets.PUBACK {
		return errors.New("error: Didn't receive PUBACK from server")
	}

	return nil
}

// TODO: handle the return codes
func (client *Client) SendSubscribe(topics ...packets.TopicWithQoS) error {
	controlHeader := packets.ControlHeader{Type: packets.SUBSCRIBE, Flags: 2}
	varHeader := packets.SubscribeVariableHeader{}
	packetID := getAndIncrementPacketID()
	varHeader.PacketIdentifier = packetID
	payload := packets.PacketPayload{}
	payload.RawApplicationMessage = make([]byte, 0, 2*len(topics))

	for _, topicWQos := range topics {
		if topicWQos.QoS > 2 {
			return errors.New("error: impossible QoS level provided")
		}
		encodedTopic, _, err := packets.EncodeUTFString(topicWQos.Topic)
		if err != nil {
			return err
		}
		payload.RawApplicationMessage = append(payload.RawApplicationMessage, encodedTopic...)
		payload.RawApplicationMessage = append(payload.RawApplicationMessage, topicWQos.QoS)
	}

	packet := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	encodedPacket, err := packets.EncodeSubscribe(packet)
	if err != nil {
		return err
	}
	if client.BrokerConnection == nil {
		return errors.New("error: connection is closed")
	}
	_, err = (*client.BrokerConnection).Write(encodedPacket)
	if err != nil {
		return err
	}
	subackArr := client.waitingPackets.GetOrWait(packetID)
	suback, _, _ := packets.DecodePacket(*subackArr)

	if suback.ControlHeader.Type != packets.SUBACK {
		return errors.New("error: Our SUBACK got nabbed")
	}

	return nil
}

func (client *Client) SendUnsubscribe(topics ...string) error {
	controlHeader := packets.ControlHeader{Type: packets.UNSUBSCRIBE}
	varHeader := packets.UnsubscribeVariableHeader{}
	packetID := getAndIncrementPacketID()
	varHeader.PacketIdentifier = packetID
	payload := packets.PacketPayload{}
	payload.TopicList = packets.ConvertStringsToTopicsWithQos(topics...)
	packet := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	encodedPacket, err := packets.EncodeUnsubscribe(packet)
	if err != nil {
		return err
	}
	if client.BrokerConnection == nil {
		return errors.New("error: connection is closed")
	}
	_, err = (*client.BrokerConnection).Write(encodedPacket)
	if err != nil {
		return err
	}
	unsubackArr := client.waitingPackets.GetOrWait(packetID)
	suback, _, _ := packets.DecodePacket(*unsubackArr)

	if suback.ControlHeader.Type != packets.UNSUBACK {
		return errors.New("error: Our UNSUBACK got nabbed")
	}

	return nil
}

func (client *Client) SendDisconnect() error {
	controlHeader := packets.ControlHeader{}
	controlHeader.Flags = 0
	controlHeader.Type = packets.DISCONNECT
	controlHeader.RemainingLength = 0

	disconnectArr := packets.EncodeFixedHeader(controlHeader)

	if client.BrokerConnection == nil {
		return errors.New("error: connection is closed")
	}

	_, err := (*client.BrokerConnection).Write(disconnectArr)
	if err != nil {
		return err
	}
	return nil
}
