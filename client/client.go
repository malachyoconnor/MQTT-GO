package client

import (
	"MQTT-GO/network"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	WaitingPacketBufferSize = 100
	ConnectionType          = network.QUIC
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

// var (
// 	connected = atomic.Int64{}
// )

func CreateAndConnectClient(ip string, port int) (*Client, error) {
	client := CreateClient()
	err := client.SetClientConnection(ip, port)
	if err != nil {
		return nil, err
	}
	err = client.SendConnect()
	// fmt.Println("\r CONNECTED OKAY", connected.Add(1))
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
		structures.Println("Sending DISCONNECT")
		err := client.SendDisconnect()
		if err != nil {
			structures.Println("Error while disconnecting:", err)
		}

		if client.BrokerConnection != nil {
			time.Sleep(time.Millisecond * 1000)
			err = client.BrokerConnection.Close()
			if err != nil {
				structures.Println("Error while closing connection:", err)
			}
			log.Println("\nConnection closed, goodbye")
		}
	} else {
		log.Println("Client is already nil when we tried to exit")
	}

	os.Exit(0)
}
