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

	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

type ClientMessage struct {
	ClientID         *ClientID
	ClientConnection *net.Conn
	Packet           []byte
}

func CreateClientMessage(clientID ClientID, clientConnection *net.Conn, packet []byte) ClientMessage {
	clientMessage := ClientMessage{
		ClientID:         &clientID,
		ClientConnection: clientConnection,
		Packet:           packet,
	}
	return clientMessage
}

func ClientHandler(connection *net.Conn, packetPool chan<- ClientMessage, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToSubscribers, connectedClient *string, connectedClientMutex *sync.Mutex) {
	newClient, err := handleInitialConnect(connection, clientTable, packetPool)
	if err != nil {
		log.Printf("- Error handling connect from %v: %v\n", newClient.NetworkConnection.RemoteAddr(), err)
		structures.Printf("Error handling connect from %v: %v\n", newClient.NetworkConnection.RemoteAddr(), err)
		if err.Error() == "error: Client already exists" {
			connack := packets.CreateConnACK(false, 2)
			_, err := (*connection).Write(connack)
			if err != nil {
				(*connection).Close()
				structures.Println("Error while writing", err)
			} else {
				// Sleep for 50 millisconds while they digest this news that they're being disconnected before closing the connection
				time.Sleep(time.Millisecond * 50)
				(*connection).Close()
			}
		}
		connectedClientMutex.Lock()
		*connectedClient = ""
		connectedClientMutex.Unlock()
		return
	}

	log.Printf("+ Client '%v' joined from  and global addr '%v'\n", newClient.ClientIdentifier, (*connection).RemoteAddr())
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

	if err != nil {
		ServerPrintln("Error decoding clientID")
		return
	}

	reader := bufio.NewReader(*connection)
	for {
		packet, err := packets.ReadPacketFromConnection(reader)

		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			break
		}
		if err != nil {
			if !strings.HasSuffix(err.Error(), "reset_stream") {
				structures.Println("Error while reading", err)
			} else {
				structures.Println("Stream closed")
			}
			break
		}

		structures.PrintCentrally(fmt.Sprintln("RECEIVED", packets.PacketTypeName(packets.GetPacketType(packet))))
		toSend := ClientMessage{ClientID: &clientID, Packet: packet, ClientConnection: connection}
		packetPool <- toSend
	}

}

// handleInitialConnect decodes the packet to find a ClientID - if none exists
// we create one and then push the connect to be handled by message Handler
func handleInitialConnect(connection *net.Conn, clientTable *structures.SafeMap[ClientID, *Client], packetPool chan<- ClientMessage) (*Client, error) {
	buffer := make([]byte, 300)
	packetLen, err := (*connection).Read(buffer)

	packet := make([]byte, packetLen)
	copy(packet, buffer[:packetLen])

	if err != nil {
		return &Client{}, err
	}

	connectPacket, err := packets.DecodeConnect(packet)
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

	clientMsg := CreateClientMessage(clientID, connection, packet)
	packetPool <- clientMsg

	return newClient, nil
}

func handleDisconnect(client Client, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToSubscribers, connectedClient *string) {
	*connectedClient = ""

	// If the client has already been disconnected elsewhere
	// by a call to client.Disconnect
	if !clientTable.Contains(client.ClientIdentifier) || client.NetworkConnection == nil {
		return
	}

	client.Disconnect(topicToClient, clientTable)
}
