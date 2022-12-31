package clients

import (
	"MQTT-GO/structures"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type ClientID string
type ClientTable map[ClientID]*Client

type Client struct {
	// Should be a max of 23 characters!
	ClientIdentifier ClientID
	Topics           *structures.LinkedList[Topic]
	TCPConnection    net.Conn
	Queue            ClientQueue
}

func CreateClient(clientID ClientID, conn *net.Conn) Client {

	client := Client{}
	client.ClientIdentifier = clientID
	client.TCPConnection = *conn

	waitingList := make(chan struct{}, 1)
	client.Queue = ClientQueue{
		WorkBeingDone: atomic.Bool{},
		WaitingList:   &waitingList,
	}

	return client
}

func (client *Client) AddTopic(newTopic Topic) {

	if client.Topics == nil {
		newLL := structures.CreateLinkedList[Topic]()
		client.Topics = &newLL
	}

	if !client.Topics.Contains(newTopic) {
		client.Topics.Append(newTopic)
	}

}

func (client *Client) disconnectClient(topicClientMap TopicToClient) {

}

var numClientsMutex sync.Mutex
var numClients int64 = 0

func generateClientID() ClientID {
	// TODO: Make this return a better unique string per new client
	numClientsMutex.Lock()
	numClients += 1
	username := fmt.Sprint(numClients)
	numClientsMutex.Unlock()

	return ClientID(username)

}
