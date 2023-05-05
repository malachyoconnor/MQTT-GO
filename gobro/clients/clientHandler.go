package clients

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"MQTT-GO/network"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// ClientMessage is a struct that stores a client's ID, a connection to the client,
// and the packet to handle.
// This is used to pass information from the server to the message handler.
// Likewise, it is used to pass information from the message handler to the message sender.
type ClientMessage struct {
	ClientID         *ClientID
	ClientConnection network.Conn
	Packet           []byte
	OutputWaitGroup  *sync.WaitGroup
}

// CreateClientMessage creates a new ClientMessage with the given ID, connection, and packet
func CreateClientMessage(clientID ClientID, clientConnection network.Conn, packet []byte) ClientMessage {
	clientMessage := ClientMessage{
		ClientID:         &clientID,
		ClientConnection: clientConnection,
		Packet:           packet,
	}
	return clientMessage
}

// ClientHandler is a function that handles a client's connection.
// It handles the initial connect, and then listens for all packets from that client,
// and passes them to the message handler.
func ClientHandler(connection network.Conn, packetHandleChan chan<- ClientMessage,
	clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicTrie,
	connectedClient *string, connectedClientMutex *sync.Mutex) {
	newClient, err := handleInitialConnect(connection, clientTable, packetHandleChan)
	if err != nil {
		if newClient.NetworkConnection == nil {
			fmt.Println("Connection Closed before finishing connection")
			newClient.Disconnect(topicToClient, clientTable)
			return
		}
		log.Printf("- Error handling connect from %v: %v\n", newClient.NetworkConnection.RemoteAddr(), err)
		structures.Printf("Error handling connect from %v: %v\n", newClient.NetworkConnection.RemoteAddr(), err)
		if err.Error() == "error: Client already exists" {
			connack := packets.CreateConnACK(false, 2)
			_, err = connection.Write(connack)
			if err != nil {
				connection.Close()
				fmt.Println("Error while writing", err)
			} else {
				// Sleep for 50 millisconds while they digest this news that they're
				// being disconnected before closing the connection
				time.Sleep(time.Millisecond * 50)
				connection.Close()
			}
		}
		connectedClientMutex.Lock()
		*connectedClient = ""
		connectedClientMutex.Unlock()
		return
	}

	log.Printf("+ Client '%v' joined from  and global addr '%v'\n", newClient.ClientIdentifier, connection.RemoteAddr())
	// We wait 1 seconds to wait for everything else to catch up
	defer handleDisconnect(*newClient, clientTable, topicToClient, connectedClient)

	clientID := newClient.ClientIdentifier
	connectedClientMutex.Lock()
	(*connectedClient) = string(clientID)
	connectedClientMutex.Unlock()

	defer func() {
		log.Printf("+ Client %v connection closed\n", clientID)
		connectedClientMutex.Lock()
		*connectedClient = ""
		connectedClientMutex.Unlock()
	}()

	reader := bufio.NewReader(connection)
	for {
		packet, err := packets.ReadPacketFromConnection(reader)

		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			break
		}
		if err != nil {
			if !strings.HasSuffix(err.Error(), "reset_stream") {
				fmt.Println("Error while reading", err)
			} else {
				fmt.Println("Stream closed")
			}
			break
		}
		toSend := ClientMessage{ClientID: &clientID, Packet: packet, ClientConnection: connection}
		packetHandleChan <- toSend
	}

}

// handleInitialConnect decodes the packet to find a ClientID - if none exists
// we create one and then push the connect to be handled by message Handler
func handleInitialConnect(connection network.Conn, clientTable *structures.SafeMap[ClientID, *Client],
	packetPool chan<- ClientMessage) (*Client, error) {
	firstPacket := make([]byte, 300)
	packetLen, err := connection.Read(firstPacket)
	firstPacket = firstPacket[:packetLen]

	if err != nil {
		return &Client{}, err
	}

	connectPacket, err := packets.DecodeConnect(firstPacket)
	if err != nil {
		ServerPrintln("Error during initial connect", err)
		return &Client{}, err
	}

	clientID := ClientID(connectPacket.Payload.ClientID)

	if connectPacket.Payload.ClientID == "" {
		clientID = generateClientID()
	}

	newClient := CreateClient(clientID, connection)
	if clientTable.Contains(clientID) {
		return clientTable.Get(clientID), errors.New("error: Client already exists")
	}
	clientTable.Put(clientID, newClient)

	clientMsg := CreateClientMessage(clientID, connection, firstPacket)
	packetPool <- clientMsg

	return newClient, nil
}

func handleDisconnect(client Client, clientTable *structures.SafeMap[ClientID, *Client],
	topicToClient *TopicTrie, connectedClient *string) {
	*connectedClient = ""

	// If the client has already been disconnected elsewhere
	// by a call to client.Disconnect
	if !clientTable.Contains(client.ClientIdentifier) || client.NetworkConnection == nil {
		return
	}

	client.Disconnect(topicToClient, clientTable)
}
