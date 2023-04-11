package gobro

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"MQTT-GO/gobro/clients"
	"MQTT-GO/network"
	"MQTT-GO/structures"
)

var (
	ServerIP          = flag.String("serverip", "localhost", "Address to host on")
	ServerPort        = flag.Int("serverport", 8000, "Port to listen on")
	ScheduledShutdown = flag.Float64("shutdown", 0.0, "Schedule a shutdown after a certain number of hours")
	ConnectionType    = network.TCP
)

type Server struct {
	clientTable    *structures.SafeMap[clients.ClientID, *clients.Client]
	topicClientMap *clients.TopicToSubscribers
	inputChan      *chan clients.ClientMessage
	outputChan     *chan clients.ClientMessage
	logFile        *os.File
}

func NewServer() Server {

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
	cleanupAndExit(server)
}

func (server *Server) StartServer() {
	flag.Parse()
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	server.logFile = file
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	go scheduleShutdown(server)

	listenForExit(server)
	log.SetOutput(file)
	// Sets the log to storefile & line numbers
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.Println("--Server starting--")
	// Listen for connections
	clients.ServerPrintln("Listening for connections via", []string{"TCP", "QUIC", "UDP"}[ConnectionType])
	clients.ServerPrintf("Listening on %v\n", getServerIpAndPort())

	listener, err := network.NewListener(ConnectionType)
	if err != nil {
		log.Print(err)
		clients.ServerPrintln("FATAL:", err)
		return
	}
	err = listener.Listen(*ServerIP, *ServerPort)

	if err != nil {
		log.Println("- Error while trying to listen for TCP connections:", err)
		clients.ServerPrintln("Error while trying to listen for TCP connections:", err)
		return
	}
	defer listener.Close()

	msgSender := CreateMessageSender(server.outputChan)
	go msgSender.ListenAndSend(server)

	msgListener := CreateMessageHandler(server.inputChan, server.outputChan)
	go msgListener.Listen(server)

	AcceptConnections(listener, server)
}

var (
	connectedClients      = make([]string, 100)
	connectedClientsMutex = sync.Mutex{}
)

func AcceptConnections(listener network.Listener, server *Server) {
	clients.ServerPrintln("Connected clients:", connectedClients)
	waitingToPrint := sync.Mutex{}
	lastPrintTime := time.Now()
	for {
		connection, err := (listener).Accept()
		if err != nil {
			log.Printf("Error while accepting a connection from '%v': %v\n", connection.RemoteAddr(), err)
			return
		}

		clients.ServerPrintln("Accepted a connection")
		// // Set a keep alive period because there isn't a foolproof way of checking if the connection
		// // suddenly closes - we want to wait for DISCONNECT messages or timeout.
		// err = connection.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second)

		if err != nil {
			log.Println("- Error while setting a keep alive period:", err)
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
			if time.Since(lastPrintTime) < time.Millisecond*500 {
				return
			}
			// Prevent output writing overtop of itself
			if !waitingToPrint.TryLock() {
				return
			}
			lastPrintTime = time.Now()
			clients.ServerPrintf("Connected clients: ")
			structures.PrintArray(connectedClients, "")
			waitingToPrint.Unlock()
		}()

		go clients.ClientHandler(&connection, *server.inputChan, server.clientTable, server.topicClientMap, newArrayPos)

	}
}

func scheduleShutdown(server *Server) {

	if *ScheduledShutdown != 0 {
		// Convert from microseconds to seconds to hours
		time.Sleep(time.Duration(*ScheduledShutdown * 1000000000 * 3600))
		server.StopServer()
	}

}

func listenForExit(server *Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			cleanupAndExit(server)
		}
	}()
}

func cleanupAndExit(server *Server) {
	for _, client := range server.clientTable.Values() {
		client.Disconnect(server.topicClientMap, server.clientTable)
	}
	server.topicClientMap.DeleteAll()
	log.Print("--Server exiting--\n\n")
	if server.logFile != nil {
		server.logFile.Close()
	}
	os.Exit(0)
}

func getServerIpAndPort() string {
	return fmt.Sprint(*ServerIP, ":", *ServerPort)
}

func DisableStdOutput() {
	clients.VerboseOutput = false
}

func EnableStdOutput() {
	clients.VerboseOutput = true
}
