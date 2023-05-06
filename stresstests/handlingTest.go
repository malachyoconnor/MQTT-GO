package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/gobro/clients"
	"MQTT-GO/structures"
	"fmt"
	"os"
	"time"
)

func TestHandleTime(numPackets int, ip string, port int, packetSize int, numberOfClients int) {
	fmt.Println("Testing Latency")

	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	location := fmt.Sprint("data/handlingTest/", transportProtocol, "/clients_", numberOfClients, "/")
	os.MkdirAll(location, os.ModePerm)
	go structures.WriteToCsv(fmt.Sprint(location, "packet_", packetSize, "B", ".csv"))
	defer structures.StopWriting()

	fmt.Println("Transport protocol:", transportProtocol)
	client.LogLatency = true
	clients.LogLatency = true

	server := gobro.NewServer()
	go server.StartServer("localhost", 80)
	defer server.StopServer(true)
	time.Sleep(1500 * time.Millisecond)

	ManyClientsPublish(ip, port, packetSize, numberOfClients)

}
