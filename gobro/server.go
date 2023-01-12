package gobro

import (
	"fmt"
	"net"
	"sync"
	"time"

	"MQTT-GO/gobro/clients"
	"MQTT-GO/structures"
)

var (
	ADDRESS    = "localhost"
	PORT       = "8000"
	serverStop = make(chan struct{}, 1)
)

type Server struct {
	clientTable    *structures.SafeMap[clients.ClientID, *clients.Client]
	topicClientMap *clients.TopicToSubscribers
	inputChan      *chan clients.ClientMessage
	outputChan     *chan clients.ClientMessage
}

func CreateServer() Server {
	clientTable := clients.CreateClientTable()
	topicClientMap := clients.CreateTopicMap()
	inputChan := make(chan clients.ClientMessage)
	outputChan := make(chan clients.ClientMessage)

	return Server{
		clientTable:    clientTable,
		topicClientMap: topicClientMap,
		inputChan:      &inputChan,
		outputChan:     &outputChan,
	}
}

func (server *Server) StopServer() {
	serverStop <- struct{}{}
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

	AcceptConnections(&listener, server)
}

var (
	connectedClients      = make([]string, 100)
	connectedClientsMutex = sync.Mutex{}
)

func AcceptConnections(listener *net.Listener, server *Server) {
	fmt.Println("??", connectedClients)
	for {

		if len(serverStop) != 0 {
			fmt.Println("Stopping server")
			return
		}

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
		connectedClientsMutex.Lock()
		for i, val := range connectedClients {
			if val == "" {
				newArrayPos = &(connectedClients[i])
				break
			}
			if i == len(connectedClients)-1 {
				connectedClients = append(connectedClients, "")
				newArrayPos = &connectedClients[len(connectedClients)-1]
			}
		}
		// Done so that another thread doesn't also use this array position
		*newArrayPos = "taken"
		connectedClientsMutex.Unlock()

		go func() {
			time.Sleep(time.Millisecond * 200)
			fmt.Print("Connected clients: ")
			structures.PrintArray(connectedClients, "")
		}()

		go clients.ClientHandler(&connection, *server.inputChan, server.clientTable, server.topicClientMap, newArrayPos)

	}
}
