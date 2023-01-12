package clients

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"MQTT-GO/structures"
)

type ClientID string

// Topics is used so we don't have to search the whole topic tree when removing a client
type Client struct {
	ClientIdentifier ClientID
	Topics           *structures.LinkedList[Topic]
	TCPConnection    net.Conn
	Tickets          *structures.TicketStand
}

func CreateClient(clientID ClientID, conn *net.Conn) *Client {
	client := Client{}
	client.ClientIdentifier = clientID
	client.TCPConnection = *conn
	client.Tickets = structures.CreateTicketStand()

	return &client
}

func (client *Client) AddTopic(newTopic Topic) {
	if client.Topics == nil {
		newLL := structures.CreateLinkedList[Topic]()
		client.Topics = newLL
	}

	if !client.Topics.Contains(newTopic) {
		client.Topics.Append(newTopic)
	}
}

func (client *Client) RemoveTopic(newTopic Topic) error {
	if client.Topics == nil {
		return errors.New("error: Client has not initialized a topic list")
	}
	return client.Topics.Delete(newTopic)
}

func (client *Client) Disconnect(topicClientMap *TopicToSubscribers, clientTable *structures.SafeMap[ClientID, *Client]) {
	if client == nil {
		return
	}
	// If the client has subscribed to something we need to remove that client
	// from the topic to client lists for each topic
	topicClientMap.DeleteClientSubscriptions(client)
	clientTable.Delete(client.ClientIdentifier)
	client.Topics.DeleteLinkedList()
	client.TCPConnection.Close()
}

var (
	numClientsMutex sync.Mutex
	numClients      int64 = 0
)

func generateClientID() ClientID {
	// TODO: Make this return a better unique string per new client
	numClientsMutex.Lock()
	numClients += 1
	username := fmt.Sprint(numClients)
	numClientsMutex.Unlock()

	return ClientID(username)
}
