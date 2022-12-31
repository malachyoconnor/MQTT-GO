package gobro

import (
	"MQTT-GO/clients"
	"fmt"
	"net"
	"time"
)

var (
	ADDRESS = "localhost"
	PORT    = "8000"
)

type Server struct {
	clientTable    *clients.ClientTable
	topicClientMap *clients.TopicToClient
	inputChan      *chan clients.ClientMessage
	outputChan     *chan clients.ClientMessage
}

func CreateServer() Server {

	clientTable := make(clients.ClientTable)
	topicClientMap := make(clients.TopicToClient)
	inputChan := make(chan clients.ClientMessage)
	outputChan := make(chan clients.ClientMessage)

	return Server{
		clientTable:    &clientTable,
		topicClientMap: &topicClientMap,
		inputChan:      &inputChan,
		outputChan:     &outputChan,
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
	connectedClients := [50]string{""}
	fmt.Println("??", connectedClients)
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

		var newArrayPos *string
		for i, val := range connectedClients {
			if val == "" {
				newArrayPos = &(connectedClients[i])
				break
			}
		}

		fmt.Println("Connected clients before new client: ", connectedClients)

		go clients.ClientHandler(&connection, server.inputChan, server.clientTable, newArrayPos)

	}
}

// msgListener := gobro.CreateMessageHandler(packetPool)
// msgListener2 := gobro.CreateMessageHandler(packetPool)

// go gobro.Sniff(packetPool)
// go msgListener.Listen()
// go msgListener2.Listen()
