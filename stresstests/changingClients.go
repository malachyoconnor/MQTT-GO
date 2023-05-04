package stresstests

import (
	"MQTT-GO/gobro"
	"MQTT-GO/network"
	"MQTT-GO/structures"
	"flag"
	"fmt"
	"time"
)

var (
	numClients     = flag.Int("clients", 100, "Profile code, and write that profile to a file")
	ConnectionType = network.TCP
)

func TestManyClients() {
	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	fmt.Println("Transport protocol:", transportProtocol)
	location := fmt.Sprint("data/messageSize/", transportProtocol, "/")

	server := gobro.NewServer()
	go server.StartServer(*ip, 6000)
	time.Sleep(time.Second)
	go structures.WriteToCsv(fmt.Sprint(location, *numClients, "_clients.csv"))

	ManyClientsPublish("localhost", 6000, 1024*3)
	structures.StopWriting()
	server.StopServer(false)
}
