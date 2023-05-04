// Package client is the main package that is used to create a client and connect to a broker.
// This can be run from the command line, or used as a utility to create MQTT clients
// programatically.
package client

import (
	"MQTT-GO/network"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	// ConnectionType is the type of transport protocol that is used
	// It is set by main.go, and can be either TCP, UDP or QUIC
	ConnectionType     = network.TCP
	PublishToWildcards = false
)

// Client is the main struct that is used to create a client and connect to a broker.
// It stores the ClientID, the connection to the broker, a buffer for incmoing packets,
// and a list of packets that are waiting for an ACK.
type Client struct {
	ClientID         string
	BrokerConnection network.Conn
	ReceivedPackets  structures.LinkedList[*packets.Packet]
	WaitingAckStruct *WaitingAcks
}

// CreateClient creates a new client with a random ClientID, and a buffer for incoming packets.
func CreateClient() *Client {
	waitingPackets := CreateWaitingAckList()
	return &Client{
		ReceivedPackets:  *structures.CreateLinkedList[*packets.Packet](),
		ClientID:         generateRandomClientID(),
		WaitingAckStruct: waitingPackets,
	}
}

// CreateAndConnectClient creates a new client, sends a connect packet to the broker, and starts listening for packets.
func CreateAndConnectClient(ip string, port int) (*Client, error) {
	client := CreateClient()
	err := client.SetClientConnection(ip, port)

	if err != nil {
		return nil, err
	}
	err = client.SendConnect(ip, port)
	go client.ListenForPackets()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SetClientConnection sets the connection to the broker.
func (client *Client) SetClientConnection(ip string, port int) error {
	connection, err := network.NewConn(ConnectionType)
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
	if client == nil {
		log.Println("Client is already nil when we tried to exit")
		os.Exit(0)
	}

	fmt.Println("Sending DISCONNECT")
	err := client.SendDisconnect()
	if err != nil {
		fmt.Println("Error while disconnecting:", err)
	}

	if client.BrokerConnection != nil {
		time.Sleep(time.Millisecond * 500)
		err = client.BrokerConnection.Close()
		if err != nil {
			fmt.Println("Error while closing connection:", err)
		}
		log.Println("\nConnection closed, goodbye")
	}

	os.Exit(0)
}
