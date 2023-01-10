package clients

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

type ClientMessage struct {
	ClientID         *ClientID
	ClientConnection *net.Conn
	Packet           *[]byte
}

func CreateClientMessage(clientID ClientID, clientConnection *net.Conn, packet *[]byte) ClientMessage {
	clientMessage := ClientMessage{
		ClientID:         &clientID,
		ClientConnection: clientConnection,
		Packet:           packet,
	}
	return clientMessage
}

func ClientHandler(connection *net.Conn, packetPool chan<- ClientMessage, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToSubscribers, connectedClient *string) {

	newClient, err := handleInitialConnect(connection, clientTable, packetPool)
	if err != nil {
		fmt.Println("Error handling connect ", err)
		if err.Error() == "error: Client already exists" {
			connack := packets.CreateConnACK(false, 2)
			(*connection).Write(connack)
			(*connection).Close()
		}
		*connectedClient = ""
		return
	}

	// We wait 1 seconds to wait for everything else to catch up
	defer handleDisconnect(*newClient, clientTable, topicToClient, connectedClient)

	clientID := newClient.ClientIdentifier
	(*connectedClient) = string(clientID)

	if err != nil {
		fmt.Println("Error decoding clientID")
		return
	}

	reader := bufio.NewReader(*connection)
	for {

		packet, err := packets.ReadPacketFromConnection(reader)

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error while reading", err)
			client := clientTable.Get(clientID)
			if client == nil {
				return
			}
			break
		}

		structures.PrintCentrally(fmt.Sprintln("RECEIVED", packets.PacketTypeName(packets.GetPacketType(packet))))
		fmt.Println(clientTable.Get(clientID))
		toSend := ClientMessage{ClientID: &clientID, Packet: packet, ClientConnection: connection}
		packetPool <- toSend
	}
	fmt.Println("Client", clientID, "connection closed")
	*connectedClient = ""

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
		fmt.Println("Error during initial connect", err)
		return &Client{}, err
	}

	var clientID ClientID = ClientID(connectPacket.Payload.ClientID)

	if connectPacket.Payload.ClientID == "" {
		clientID = generateClientID()
	}

	newClient := CreateClient(clientID, connection)
	if clientTable.Contains(clientID) {
		return clientTable.Get(clientID), errors.New("error: Client already exists")
	}
	clientTable.Put(clientID, newClient)

	clientMsg := CreateClientMessage(clientID, connection, &packet)
	packetPool <- clientMsg

	return newClient, nil

}

func handleDisconnect(client Client, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToSubscribers, connectedClient *string) {
	*connectedClient = ""
	fmt.Println("Client handler disconnecting client")
	// time.Sleep(3 * time.Second)

	// If the client has already been disconnected elsewhere
	// by a call to client.Disconnect
	if !clientTable.Contains(client.ClientIdentifier) {
		return
	}

	client.Disconnect(topicToClient, clientTable)
}
