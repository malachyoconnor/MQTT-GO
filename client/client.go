package client

import (
	"MQTT-GO/network"
	"MQTT-GO/packets"
	"log"
	"os"
	"os/signal"
)

var (
	WaitingPacketBufferSize = 100
	ConnectionType          = network.TCP
)

type Client struct {
	ClientID         string
	BrokerConnection network.Con
	ReceivedPackets  chan *packets.Packet
	waitingAckStruct *WaitingAcks
}

func CreateClient() *Client {
	messageChannel := make(chan *packets.Packet, WaitingPacketBufferSize)
	waitingPackets := CreateWaitingPacketList()
	return &Client{
		ReceivedPackets:  messageChannel,
		ClientID:         generateRandomClientID(),
		waitingAckStruct: waitingPackets,
	}
}

func CreateAndConnectClient(ip string, port int) (*Client, error) {
	client := CreateClient()
	err := client.SetClientConnection(ip, port)
	if err != nil {
		return nil, err
	}
	err = client.SendConnect()
	if err != nil {
		return nil, err
	}
	go client.ListenForPackets()
	return client, nil
}

func (client *Client) SetClientConnection(ip string, port int) error {
	connection, err := network.NewCon(ConnectionType)
	if err != nil {
		return err
	}
	err = connection.Connect(ip, port)
	if err != nil {
		return err
	}

	client.BrokerConnection = connection
	return nil
}

func listenForExit(client *Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			cleanupAndExit(client)
		}
	}()
}

func cleanupAndExit(client *Client) {
	if client != nil {
		client.SendDisconnect()
		if client.BrokerConnection != nil {
			client.BrokerConnection.Close()
			log.Println("\nConnection closed, goodbye")
		}
	} else {
		log.Println("Client is already nil when we tried to exit")
	}

	os.Exit(0)
}
