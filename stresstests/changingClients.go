package stresstests

import (
	"MQTT-GO/gobro"
	"MQTT-GO/network"
	"MQTT-GO/structures"
	"fmt"
	"os"
	"time"
)

var (
	ConnectionType = network.TCP
)

func TestManyClients(numClients int, ip string, port int, packetSize int) {
	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	fmt.Println("Transport protocol:", transportProtocol)

	location := fmt.Sprint("data/messageSize/", packetSize, "/", transportProtocol, "/")
	err := os.MkdirAll(location, os.ModePerm)

	if err != nil {
		panic(err)
	}

	server := gobro.NewServer()
	go server.StartServer("localhost", 80)
	time.Sleep(time.Second)
	go structures.WriteToCsv(fmt.Sprint(location, numClients, "_clients.csv"))

	ManyClientsPublish(ip, port, packetSize, numClients)
	structures.StopWriting()
	server.StopServer(false)
}
