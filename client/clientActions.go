package client

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"MQTT-GO/network"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// Here we'll have the functions that make the client perform it's actions.

var (
	errConnectionClosed = errors.New("error: connection is closed")
)

// SendConnect encodes a connect packet and sends it to the broker.
func (client *Client) SendConnect(ip string, port int) error {
	if client.BrokerConnection == nil {
		return errors.New("error: Client does not have a broker connection")
	}

	controlHeader := packets.ControlHeader{Type: packets.CONNECT, Flags: 0}
	varHeader := packets.ConnectVariableHeader{KeepAlive: 3600}
	payload := packets.PacketPayload{}
	payload.ClientID = client.ClientID

	connectPacket := packets.CombinePacketSections(&controlHeader, &varHeader, &payload)
	connectPacketArr, err := packets.EncodeConnect(connectPacket)
	if err != nil {
		return err
	}
	if client.BrokerConnection == nil {
		return errConnectionClosed
	}
	_, err = client.BrokerConnection.Write(connectPacketArr)

	if err != nil {
		return err
	}

	var result []byte

	readPacketChannel := make(chan []byte, 1)
	go func() {
		buffer := make([]byte, 1024*3)
		n, _ := client.BrokerConnection.Read(buffer)
		readPacketChannel <- buffer[:n]
	}()

	for {
		select {
		case result = <-readPacketChannel:
			{
				break
			}
		case <-getTimeoutChannel(1 * time.Second):
			{
				_, err = client.BrokerConnection.Write(connectPacketArr)
				continue
			}
		}
		break
	}

	structures.Println(result, "Read connack")
	if err != nil {
		return err
	}
	packet, _, _ := packets.DecodePacket(result)

	if packet.ControlHeader.Type != packets.CONNACK {
		structures.PrintInterface(packet)
		return errors.New("error: Received packet other than CONNACK from server")

	} else if packet.VariableLengthHeader.(*packets.ConnackVariableHeader).ConnectReturnCode == 2 {
		fmt.Println("      ClientID already exists, generating new one...")
		// If the clientID already exists then we wait
		time.Sleep(time.Millisecond)
		client.ClientID = generateRandomClientID()
		client.BrokerConnection.Close()
		err := client.SetClientConnection(ip, port)
		if err != nil {
			fmt.Println("Error while setting client connection")
			return err
		}
		return client.SendConnect(ip, port)
	}

	return nil
}

func getTimeoutChannel(timeout time.Duration) chan struct{} {
	timeoutChannel := make(chan struct{}, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChannel <- struct{}{}
	}()
	return timeoutChannel
}

// TODO: Handle readPacketFromConnection error properly
// TODO: Check if everything you would need for a publish packet is present!

// SendPublish encodes a publish packet and sends it to the broker.
func (client *Client) SendPublish(applicationMessage []byte, topic string) error {
	// If the topic contains wildcards and we don't want to publish to wildcards then return an error
	if !PublishToWildcards && (strings.Contains(topic, "+") || strings.Contains(topic, "#")) {
		return errors.New("error: Cannot publish to topics with wildcards + or #")
	}

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
		return errConnectionClosed
	}

	n, err := (client.BrokerConnection).Write(publishPacketArr)

	if LogLatency {
		SendingLatencyChannel <- &network.LatencyStruct{T: time.Now(), PacketID: packetID}
	}

	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("error: Wrote 0 bytes to connection")
	}

	if controlHeader.Flags&6 == 0 {
		return nil
	}

	pubackArr := client.WaitingAckStruct.GetOrWait(packetID)
	packet, _, _ := packets.DecodePacket(*pubackArr)

	if packet.ControlHeader.Type != packets.PUBACK {
		return errors.New("error: Didn't receive PUBACK from server")
	}

	return nil
}

// TODO: handle the return codes

// SendSubscribe encodes a subscribe packet and sends it to the broker.
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
		return errConnectionClosed
	}
	_, err = client.BrokerConnection.Write(encodedPacket)
	if err != nil {
		return err
	}
	subackArr := client.WaitingAckStruct.GetOrWait(packetID)
	suback, _, _ := packets.DecodePacket(*subackArr)

	if suback.ControlHeader.Type != packets.SUBACK {
		return errors.New("error: Our SUBACK got nabbed")
	}

	return nil
}

// SendUnsubscribe encodes an unsubscribe packet and sends it to the broker.
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
		return errConnectionClosed
	}
	_, err = client.BrokerConnection.Write(encodedPacket)
	if err != nil {
		return err
	}
	unsubackArr := client.WaitingAckStruct.GetOrWait(packetID)
	suback, _, _ := packets.DecodePacket(*unsubackArr)

	if suback.ControlHeader.Type != packets.UNSUBACK {
		return errors.New("error: Our UNSUBACK got nabbed")
	}

	return nil
}

// SendDisconnect encodes a disconnect packet and sends it to the broker.
func (client *Client) SendDisconnect() error {
	controlHeader := packets.ControlHeader{}
	controlHeader.Flags = 0
	controlHeader.Type = packets.DISCONNECT
	controlHeader.RemainingLength = 0
	disconnectArr := packets.EncodeFixedHeader(controlHeader)

	if client.BrokerConnection == nil {
		return errConnectionClosed
	}

	n, err := client.BrokerConnection.Write(disconnectArr)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("error: No bytes written")
	}
	return nil
}

// SendDisconnect encodes a SUBACK packet and sends it to the broker.
func (client *Client) SendSUBACK() error {
	controlHeader := packets.ControlHeader{}
	controlHeader.Flags = 0
	controlHeader.Type = packets.SUBACK
	controlHeader.RemainingLength = 0
	subackArr := packets.EncodeFixedHeader(controlHeader)

	if client.BrokerConnection == nil {
		return errConnectionClosed
	}

	structures.Println("Sending SubAck")
	n, err := client.BrokerConnection.Write(subackArr)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("error: No bytes written")
	}
	return nil
}

func (client *Client) SendPuback(packetID int) error {
	if client.BrokerConnection == nil {
		return errConnectionClosed
	}

	toSend := packets.CreatePubAck(packetID)

	n, err := client.BrokerConnection.Write(toSend)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("error: No bytes written")
	}
	return nil
}
