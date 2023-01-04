package gobro

import (
	"MQTT-GO/clients"
	"MQTT-GO/structures"
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

	clientTable := clients.CreateClientTable()
	topicClientMap := make(clients.TopicToClient)
	inputChan := make(chan clients.ClientMessage)
	outputChan := make(chan clients.ClientMessage)

	return Server{
		clientTable:    clientTable,
		topicClientMap: &topicClientMap,
		inputChan:      &inputChan,
		outputChan:     &outputChan,
	}

}

func (server *Server) StartServer() {

	// Listen for TCP connections
	fmt.Println("Listening for TCP connections")
	listener, err := net.Listen("tcp", ADDRESS+":"+PORT)

	if err != nil {
		fmt.Println("Error while trying to listen for TCP connections", err)
		return
	}
	defer listener.Close()

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

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Accepted a connection")
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

		go func() {
			time.Sleep(time.Millisecond * 200)
			fmt.Print("Connected clients: ")
			structures.PrintArray(connectedClients[:], "")
		}()

		go clients.ClientHandler(&connection, server.inputChan, server.clientTable, server.topicClientMap, newArrayPos)

	}
}
