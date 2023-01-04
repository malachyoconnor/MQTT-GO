package clients

import (
	"MQTT-GO/structures"
	"fmt"
	"net"
	"sync"
)

type ClientID string

type Client struct {
	// Should be a max of 23 characters!
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
		client.Topics = &newLL
	}

	if !client.Topics.Contains(newTopic) {
		client.Topics.Append(newTopic)
	}

}

func (client *Client) Disconnect(topicClientMap *TopicToClient, clientTable *ClientTable) {
	if client == nil {
		return
	}

	// If the client has subscribed to something we need to
	// remove that client from the topic to client lists for each
	// topic
	if client.Topics != nil {
		node := client.Topics.Head()
		for node != nil {
			(*topicClientMap)[node.Value()].PrintItems()
			err := (*topicClientMap)[node.Value()].Delete(client.ClientIdentifier)
			(*topicClientMap)[node.Value()].PrintItems()
			if err != nil {
				panic("Attempted to remove client from Topic list that isn't there")
			}
			node = node.Next()
		}
	}

	clientTable.Delete(client.ClientIdentifier)
	client.TCPConnection.Close()
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
