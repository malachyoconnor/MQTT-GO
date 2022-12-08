package gobro

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"fmt"
	"net"
)

var (
	ADDRESS = "localhost"
	PORT    = "8000"
)

type Server struct {
	clientTable       *client.ClientTable
	SubscriptionTable *SubscriptionTable
	inputPool         *packets.BytePool
	outputPool        *packets.BytePool
}

func CreateServer() Server {

	clientTable := make(client.ClientTable)
	SubscriptionTable := make(SubscriptionTable)
	inputPool := packets.CreateBytePool()
	outputPool := packets.CreateBytePool()

	return Server{
		clientTable:       &clientTable,
		SubscriptionTable: &SubscriptionTable,
		inputPool:         inputPool,
		outputPool:        outputPool,
	}

}

func (server *Server) StartServer() {
	// Listen for TCP connections
	listener, err := net.Listen("tcp", ADDRESS+":"+PORT)
	defer listener.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	go AcceptConnections(&listener, server)

	msgSender := CreateMessageSender(server.outputPool)
	go msgSender.SendMessages(server)

	msgListener := CreateMessageHandler(server.inputPool, server.outputPool)
	go msgListener.Listen(server)

	for {
	}

}

func AcceptConnections(listener *net.Listener, server *Server) {
	for {
		connection, err := (*listener).Accept()
		defer connection.Close()
		fmt.Println("Accepted a connection")

		if err != nil {
			fmt.Println(err)
			return
		}

		go client.ClientHandler(&connection, server.inputPool)
	}
}

// msgListener := gobro.CreateMessageHandler(packetPool)
// msgListener2 := gobro.CreateMessageHandler(packetPool)

// go gobro.Sniff(packetPool)
// go msgListener.Listen()
// go msgListener2.Listen()
