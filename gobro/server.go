package gobro

import (
	"MQTT-GO/client"
	"fmt"
	"net"
	"time"
)

var (
	ADDRESS = "localhost"
	PORT    = "8000"
)

type Server struct {
	clientTable       *client.ClientTable
	SubscriptionTable *SubscriptionTable
	inputChan         *chan client.ClientMessage
	outputChan        *chan client.ClientMessage
}

func CreateServer() Server {

	clientTable := make(client.ClientTable)
	SubscriptionTable := make(SubscriptionTable)
	inputChan := make(chan client.ClientMessage)
	outputChan := make(chan client.ClientMessage)

	return Server{
		clientTable:       &clientTable,
		SubscriptionTable: &SubscriptionTable,
		inputChan:         &inputChan,
		outputChan:        &outputChan,
	}

}

func (server *Server) StartServer() {
	// Listen for TCP connections
	fmt.Println("Listening for TCP connections")
	listener, err := net.Listen("tcp", ADDRESS+":"+PORT)
	defer listener.Close()

	if err != nil {
		fmt.Println("Error while trying to listen for TCP connections", err)
		return
	}

	msgSender := CreateMessageSender(server.outputChan)
	go msgSender.ListenAndSend(server)

	msgListener := CreateMessageHandler(server.inputChan, server.outputChan)
	go msgListener.Listen(server)

	go AcceptConnections(&listener, server)

	for {
	}

}

func AcceptConnections(listener *net.Listener, server *Server) {
	for {
		connection, err := (*listener).Accept()
		fmt.Println("Accepted a connection")

		if err != nil {
			fmt.Println(err)
			return
		}

		// Set a keep alive period because there isn't a foolproof way of checking if the connection
		// suddenly closes - we want to wait for DISCONNECT messages or timeout.
		err = connection.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second)

		if err != nil {
			fmt.Println(err)
			return
		}

		go client.ClientHandler(&connection, server.inputChan, server.clientTable)
	}
}

// msgListener := gobro.CreateMessageHandler(packetPool)
// msgListener2 := gobro.CreateMessageHandler(packetPool)

// go gobro.Sniff(packetPool)
// go msgListener.Listen()
// go msgListener2.Listen()
