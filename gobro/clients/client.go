// Package clients is used to store information about and handle clients.
// It is used by the gobro package to store information about clients and differs
// from the client package.
package clients

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"MQTT-GO/structures"
)

// ClientID is a string that is used to uniquely identify a client
type ClientID string

// Client is a struct that stores a client's ID, a list of topics they are subscribed to,
// a connection to the client, and a ticket stand for sending messages in order.
type Client struct {
	ClientIdentifier  ClientID
	Topics            *structures.LinkedList[Topic]
	NetworkConnection net.Conn
	Tickets           *structures.TicketStand
}

// CreateClient creates a new client with the given ID and connection
func CreateClient(clientID ClientID, conn *net.Conn) *Client {
	client := Client{}
	client.ClientIdentifier = clientID
	client.NetworkConnection = *conn
	client.Tickets = structures.CreateTicketStand()

	return &client
}

// AddTopic adds a topic to the client's list of subscribed topics
// If the client has not initialized a topic list, it will be initialized
// If the client is already subscribed to the topic, it will not be added
func (client *Client) AddTopic(newTopic Topic) {
	if client.Topics == nil {
		newLL := structures.CreateLinkedList[Topic]()
		client.Topics = newLL
	}

	if !client.Topics.Contains(newTopic) {
		client.Topics.Append(newTopic)
	}
}

// RemoveTopic removes a topic from the client's list of subscribed topics
// If the client has not initialized a topic list, an error will be returned
// If the client is not subscribed to the topic, an error will be returned
func (client *Client) RemoveTopic(newTopic Topic) error {
	if client.Topics == nil {
		return errors.New("error: Client has not initialized a topic list")
	}
	return client.Topics.Delete(newTopic)
}

// Disconnect removes the client from the client table and removes the client from
// the topic to client map for each topic the client is subscribed to.
func (client *Client) Disconnect(topicClientMap *TopicToSubscribers,
	clientTable *structures.SafeMap[ClientID, *Client]) {
	if client == nil {
		return
	}
	// If the client has subscribed to something we need to remove that client
	// from the topic to client lists for each topic
	topicClientMap.DeleteClientSubscriptions(client)
	clientTable.Delete(client.ClientIdentifier)
	client.Topics.DeleteLinkedList()
	client.NetworkConnection.Close()
}

var (
	numClientsMutex sync.Mutex
	numClients      int64
)

func generateClientID() ClientID {
	numClientsMutex.Lock()
	numClients++
	username := fmt.Sprint(numClients)
	numClientsMutex.Unlock()

	return ClientID(username)
}

var (
	// VerboseOutput is a boolean that determines whether or not the server will print
	// verbose output to the console
	VerboseOutput = true
)

// ServerPrintln prints the given arguments to the console if VerboseOutput is true
func ServerPrintln(args ...any) {
	if VerboseOutput {
		for _, arg := range args {
			fmt.Println(arg, " ")
		}
	}
}

// ServerPrintf prints the given arguments to the console if VerboseOutput is true
func ServerPrintf(format string, args ...any) {
	if VerboseOutput {
		fmt.Printf(format, args...)
	}
}
