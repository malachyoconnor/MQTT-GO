package clients

import (
	"MQTT-GO/packets"
	"bufio"
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

func handleDisconnect(clientTable *ClientTable, clientID ClientID) {

	fmt.Println("Disconnecting client", clientID)
	delete(*clientTable, clientID)

}

func ClientHandler(connection *net.Conn, packetPool *chan ClientMessage, clientTable *ClientTable, connectedClient *string) {
	defer func() {
		if *connection != nil {
			(*connection).Close()
		}
		*connectedClient = ""
	}()

	newClient, err := handleInitialConnect(connection, clientTable, packetPool)
	if err != nil {
		fmt.Println("Error handling connect ", err)
		return
	}
	clientID := newClient.ClientIdentifier
	fmt.Println(*connectedClient)
	*connectedClient = string(clientID)

	if err != nil {
		fmt.Println("Error decoding clientID")
		return
	}

	reader := bufio.NewReader(*connection)
	for {
		buffer, err := reader.Peek(4)

		if (err != nil) && len(buffer) == 0 {
			client := (*clientTable)[clientID]
			if client == nil {
				return
			}

			time.Sleep(time.Second)
			client.Queue.JoinWaitList()
			// We wait 1 seconds to wait for everything else to catch up
			handleDisconnect(clientTable, clientID)
			break
		}

		dataLen, varLengthIntLen, err := packets.DecodeVarLengthInt(buffer[1:])
		packet := make([]byte, dataLen+varLengthIntLen+1)
		bytesRead, err := io.ReadFull(reader, packet)
		packet = packet[:bytesRead]

		if err != nil {
			fmt.Println("packet:", packet)
			fmt.Println("Error: ", err)
			break
		}

		toSend := ClientMessage{ClientID: &clientID, Packet: &packet, ClientConnection: connection}
		(*packetPool) <- toSend
	}

	handleDisconnect(clientTable, clientID)
	fmt.Println("Client connection closed")
}

// handleInitialConnect decodes the packet to find a ClientID - if none exists
// we create one and then push the connect to be handled by message Handler
func handleInitialConnect(connection *net.Conn, clientTable *ClientTable, packetPool *chan ClientMessage) (*Client, error) {
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

	clientMsg := CreateClientMessage(clientID, connection, &packet)
	(*packetPool) <- clientMsg

	return &newClient, nil

}
