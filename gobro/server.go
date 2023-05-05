// Package gobro is the main package that is used to create a broker and listen for clients.
// It can be run from the command line, or used as a utility to create MQTT brokers
// programatically. It uses the clients package to store information about clients,
// and the network package to handle the network connections.
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
	scheduledShutdown = flag.Float64("shutdown", 0.0, "Schedule a shutdown after a certain number of hours")
	// ConnectionType is the type of transport protocol that is used
	// It is set by main.go, and can be either TCP, UDP or QUIC
	ConnectionType = network.TCP
	PrintOutput    = false
)

// Server is the main struct that is used to create a broker and listen for clients.
// It stores a map of clients, a map of topics to subscribers, a channel for incoming packets,
// a channel for outgoing packets, and a log file.
type Server struct {
	clientTable *structures.SafeMap[clients.ClientID, *clients.Client]
	topicTrie   *clients.TopicTrie
	inputChan   *chan clients.ClientMessage
	outputChan  *chan clients.ClientMessage
	logFile     *os.File
	listener    *network.Listener
}

// NewServer creates a new server with a new client table, topic map, and channels for incoming and outgoing packets.
func NewServer() Server {
	clientTable := structures.CreateSafeMap[clients.ClientID, *clients.Client]()
	topicTrie := clients.CreateTopicTrie()

	inputChan := make(chan clients.ClientMessage, 10000)
	outputChan := make(chan clients.ClientMessage, 10000)

	return Server{
		clientTable: clientTable,
		topicTrie:   topicTrie,
		inputChan:   &inputChan,
		outputChan:  &outputChan,
	}
}

// StopServer stops the server by closing the log file and exiting the program.
func (server *Server) StopServer(shutdownProgram bool) {
	cleanupAndExit(server, shutdownProgram)
}

// StartServer starts the server by listening for connections, and then listening for packets.
// It also starts a goroutine to listen for a shutdown signal.
// It runs the msgSender and msgListener functions in separate goroutines.
// We run AcceptConnections in the main thread, because it blocks.
func (server *Server) StartServer(ip string, port int) {
	flag.Parse()
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	server.logFile = file
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	go scheduleShutdown(server)

	listenForExit(server, true)
	log.SetOutput(file)
	// Sets the log to storefile & line numbers
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.Println("--Server starting--")
	// Listen for connections
	structures.Println("Listening for connections via", []string{"TCP", "QUIC", "UDP"}[ConnectionType])
	structures.Printf("Listening on %v\n", ip+":"+fmt.Sprint(port))

	listener, err := network.NewListener(ConnectionType)
	if err != nil {
		log.Print(err)
		clients.ServerPrintln("FATAL:", err)
		return
	}
	server.listener = &listener
	err = listener.Listen(ip, port)

	if err != nil {
		log.Println("- Error while trying to listen for TCP connections:", err)
		clients.ServerPrintln("Error while trying to listen for TCP connections:", err)
		return
	}
	defer listener.Close()

	msgSender := CreateMessageSender(server.outputChan)
	go msgSender.ListenAndSend(server)
	msgHandler := CreateMessageHandler(server.inputChan, server.outputChan)
	go msgHandler.Listen(server)
	AcceptConnections(listener, server)
}

var (
	connectedClients      = make([]string, 100, 500)
	connectedClientsMutex = sync.Mutex{}
)

// AcceptConnections accepts connections from clients, and then creates a new goroutine to handle the client.
func AcceptConnections(listener network.Listener, server *Server) {
	clients.ServerPrintln("Connected clients:", connectedClients)
	go func(connectedClients []string) {
		for {
			time.Sleep(time.Second * 5)
			fmt.Println((connectedClients), "USERS")
		}
	}(connectedClients)

	waitingToPrint := sync.Mutex{}
	lastPrintTime := time.Now()
	for {
		connection, err := (listener).Accept()
		if err != nil {
			log.Printf("Error while accepting a connection")
			return
		}

		clients.ServerPrintln("Accepted a connection")
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
			connectedClientsMutex.Lock()
			structures.PrintArray(connectedClients, "")
			connectedClientsMutex.Unlock()
			waitingToPrint.Unlock()
		}()

		go clients.ClientHandler(connection, *server.inputChan, server.clientTable,
			server.topicTrie, newArrayPos, &connectedClientsMutex)
	}
}

func scheduleShutdown(server *Server) {
	if *scheduledShutdown != 0 {
		// Convert from microseconds to seconds to hours
		const hourInNano = 1000000000 * 3600
		time.Sleep(time.Duration(*scheduledShutdown * float64(hourInNano)))
		server.StopServer(true)
	}
}

func listenForExit(server *Server, exit bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("CLOSING")
			cleanupAndExit(server, exit)
		}
	}()
}

func cleanupAndExit(server *Server, exit bool) {
	for _, client := range server.clientTable.Values() {
		client.Disconnect(server.topicTrie, server.clientTable)
	}
	structures.StopWriting()
	(*server.listener).Close()
	server.topicTrie.DeleteAll()
	log.Print("--Server exiting--\n\n")
	if server.logFile != nil {
		server.logFile.Close()
	}
	if exit {
		os.Exit(0)

	}
}

// DisableStdOutput disables the output to stdout.
func DisableStdOutput() {
	clients.VerboseOutput = false
}

// EnableStdOutput enables the output to stdout.
func EnableStdOutput() {
	clients.VerboseOutput = true
}
