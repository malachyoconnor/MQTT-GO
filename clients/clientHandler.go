package clients

import (
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
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

func ClientHandler(connection *net.Conn, packetPool *chan ClientMessage, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToClient, connectedClient *string) {

	newClient, err := handleInitialConnect(connection, clientTable, packetPool)
	if err != nil {
		fmt.Println("Error handling connect ", err)
		if err.Error() == "error: Client already exists" {
			newClient.Disconnect(topicToClient, clientTable)
		}
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
		buffer, err := reader.Peek(4)

		if (err != nil) && len(buffer) == 0 {
			client := clientTable.Get(clientID)
			if client == nil {
				return
			}
			break
		}

		dataLen, varLengthIntLen, err := packets.DecodeVarLengthInt(buffer[1:])
		if err != nil {
			fmt.Println(err)
			break
		}
		packet := make([]byte, dataLen+varLengthIntLen+1)
		bytesRead, err := io.ReadFull(reader, packet)
		packet = packet[:bytesRead]

		if err != nil {
			fmt.Println("packet:", packet)
			fmt.Println("Error: ", err)
			break
		}

		structures.PrintCentrally(fmt.Sprintln("RECEIVED", packets.PacketTypeName(packets.GetPacketType(&packet))))

		toSend := ClientMessage{ClientID: &clientID, Packet: &packet, ClientConnection: connection}
		(*packetPool) <- toSend
	}
	fmt.Println("Client", clientID, "connection closed")
	*connectedClient = ""

}

// handleInitialConnect decodes the packet to find a ClientID - if none exists
// we create one and then push the connect to be handled by message Handler
func handleInitialConnect(connection *net.Conn, clientTable *structures.SafeMap[ClientID, *Client], packetPool *chan ClientMessage) (*Client, error) {
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

	var clientID ClientID = ClientID(connectPacket.Payload.ClientId)
	if connectPacket.Payload.ClientId == "" {
		clientID = generateClientID()
	}

	newClient := CreateClient(clientID, connection)
	if clientTable.Exists(clientID) {
		return clientTable.Get(clientID), errors.New("error: Client already exists")
	}
	clientTable.Put(clientID, newClient)

	clientMsg := CreateClientMessage(clientID, connection, &packet)
	(*packetPool) <- clientMsg

	return newClient, nil

}

func handleDisconnect(client Client, clientTable *structures.SafeMap[ClientID, *Client], topicToClient *TopicToClient, connectedClient *string) {
	*connectedClient = ""
	time.Sleep(3 * time.Second)

	// If the client has already been disconnected elsewhere
	// by a call to client.Disconnect
	if !clientTable.Exists(client.ClientIdentifier) {
		return
	}

	client.Disconnect(topicToClient, clientTable)
}
