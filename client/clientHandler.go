package client

import (
	"MQTT-GO/packets"
	"bufio"
	"fmt"
	"io"
	"net"
)

type ClientMessage struct {
	ClientID *string
	Packet   *[]byte
}

func ClientHandler(connection *net.Conn, packetPool *chan ClientMessage, clientTable *ClientTable) {

	defer (*connection).Close()
	newClient, err := handleInitialConnect(connection, clientTable)
	if err != nil {
		fmt.Println("Error handling connect ", err)
		return
	}
	clientID := string(newClient.ClientIdentifier)

	if err != nil {
		fmt.Println("Error decoding clientID")
		return
	}

	reader := bufio.NewReader(*connection)
	for {

		buffer, err := reader.Peek(4)

		if len(buffer) == 0 {
			continue
		}

		dataLen, varLengthIntLen, err := packets.DecodeVarLengthInt(buffer[1:])
		packet := make([]byte, dataLen+varLengthIntLen+1)
		bytesRead, err := io.ReadFull(reader, packet)
		packet = packet[:bytesRead]

		if bytesRead == 0 {
			fmt.Println(err)
			continue
		}

		toSend := ClientMessage{ClientID: &clientID, Packet: &packet}

		if err != nil {
			fmt.Println(packet)
			fmt.Println("Error: ", err)
			break
		}

		if packets.GetPacketType(&packet) == 14 {
			break
		}

		(*packetPool) <- toSend
	}

	fmt.Println("Client connection closed")
}

func handleInitialConnect(connection *net.Conn, clientTable *ClientTable) (*Client, error) {
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
		ClientIdentifier: ClientID(clientID),
		TCPConnection:    *connection,
	}

	(*clientTable)[clientID] = newClient

	// Now we send a CONNACK

	conACK := packets.CreateConnACK(true, 0)
	resultToSend := packets.EncodeConACK(&conACK)
	fmt.Println("Sending CONNACK")
	_, err = (*connection).Write(resultToSend)

	if err != nil {
		return &Client{}, err
	}

	return newClient, nil

}
