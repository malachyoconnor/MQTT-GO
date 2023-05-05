package stresstests

import (
	"MQTT-GO/gobro"
	"MQTT-GO/network"
	"MQTT-GO/structures"
	"fmt"
	"time"
)

var (
	ConnectionType = network.TCP
)

func TestManyClients(numClients int, ip string, port int) {
	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	fmt.Println("Transport protocol:", transportProtocol)
	location := fmt.Sprint("data/messageSize/", transportProtocol, "/")

	server := gobro.NewServer()
	go server.StartServer("localhost", 80)
	time.Sleep(time.Second)
	go structures.WriteToCsv(fmt.Sprint(location, numClients, "_clients.csv"))

	ManyClientsPublish(ip, port, 500, numClients)
	structures.StopWriting()
	server.StopServer(false)
}
