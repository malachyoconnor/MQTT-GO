package client

import (
	"MQTT-GO/packets"
	"bufio"
	"fmt"
	"io"
	"net"
)

type ClientMessage struct {
	ClientID         *ClientID
	ClientConnection *net.Conn
	Packet           *[]byte
}

func CreateClientMessage(clientID *ClientID, clientConnection *net.Conn, packet *[]byte) ClientMessage {
	clientMessage := ClientMessage{
		ClientID:         clientID,
		ClientConnection: clientConnection,
		Packet:           packet,
	}
	return clientMessage
}

func ClientHandler(connection *net.Conn, packetPool *chan ClientMessage, clientTable *ClientTable) {

	defer (*connection).Close()
	newClient, err := handleInitialConnect(connection, clientTable, packetPool)
	if err != nil {
		fmt.Println("Error handling connect ", err)
		return
	}
	clientID := newClient.ClientIdentifier

	if err != nil {
		fmt.Println("Error decoding clientID")
		return
	}

	reader := bufio.NewReader(*connection)
	for {

		buffer, err := reader.Peek(4)

		if err != nil {
			if err != io.EOF {
				fmt.Println("Error in ClientHandler:", err)
			}
			break
		}

		if len(buffer) == 0 {
			continue
		}

		dataLen, varLengthIntLen, err := packets.DecodeVarLengthInt(buffer[1:])
		packet := make([]byte, dataLen+varLengthIntLen+1)
		bytesRead, err := io.ReadFull(reader, packet)
		packet = packet[:bytesRead]

		if err != nil {
			fmt.Println(packet)
			fmt.Println("Error: ", err)
			break
		}

		toSend := ClientMessage{ClientID: &clientID, Packet: &packet, ClientConnection: connection}

		(*packetPool) <- toSend
	}

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

	newClient := &Client{
		// TODO: Make a function which takes a remoteAddr and makes a byte[4]
		// IPAddress:        (*connection).RemoteAddr().String(),
		ClientIdentifier: clientID,
		TCPConnection:    *connection,
	}

	clientMsg := CreateClientMessage(&clientID, *&connection, &packet)
	(*packetPool) <- clientMsg

	return newClient, nil

}
