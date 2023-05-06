package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/gobro/clients"
	"MQTT-GO/network"
	"MQTT-GO/packets"
	"fmt"
	"os"
	"time"
)

func TestLatency(numPackets int, ip string, port int, packetSize int, numberOfClients int) {
	fmt.Println("Testing Latency")

	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	fmt.Println("Transport protocol:", transportProtocol)
	client.LogLatency = true
	clients.LogLatency = true

	server := gobro.NewServer()
	go server.StartServer("localhost", 80)
	defer server.StopServer(true)
	time.Sleep(1500 * time.Millisecond)

	newClient, err := client.CreateAndConnectClient(ip, port)

	if err != nil {
		panic(err)
	}

	newClient.SendPublish([]byte("test"), "abc")
	newClient.SendSubscribe(packets.TopicWithQoS{Topic: "abc"})
	msgToSend := make([]byte, packetSize)

	publishingClients := make([]*client.Client, numberOfClients)
	connectAllClients(publishingClients, ip, port, os.Stdout)

	for i := 0; i < numPackets/numberOfClients; i++ {
		for _, c := range publishingClients {
			c.SendPublish(msgToSend, "abc")
		}

		time.Sleep(5 * time.Millisecond)
	}

	counter := 0
	for newClient.ReceivedPackets.Size() < numPackets {
		time.Sleep(100 * time.Millisecond)
		counter++
		if transportProtocol == "UDP" {
			counter += 9
		}
		if counter > 100 {
			break
		}
		fmt.Print("\rWaiting for all packets to be received", newClient.ReceivedPackets.Size(), numPackets, counter)
	}

	LogLatencies(client.SendingLatencyChannel, clients.ReceivingLatencyChannel, transportProtocol, numberOfClients, packetSize, true)
	LogLatencies(clients.SendingLatencyChannel, client.ReceivingLatencyChannel, transportProtocol, numberOfClients, packetSize, false)
}

func LogLatencies(clientChannel chan *network.LatencyStruct, gobroChannel chan *network.LatencyStruct,
	transportProtocol string, numberOfClients int, packetSize int, serverReceive bool) {

	fmt.Println(len(clientChannel), len(gobroChannel))

	clientLatencyMap, gobroLatencyMap := make(map[int]*network.LatencyStruct, len(clientChannel)), make(map[int]*network.LatencyStruct, len(gobroChannel))

	for i := 0; i < len(clientChannel); i++ {
		clientLatency := <-clientChannel
		clientLatencyMap[clientLatency.PacketID] = clientLatency
	}
	for i := 0; i < len(gobroChannel); i++ {
		gobroLatency := <-gobroChannel
		gobroLatencyMap[gobroLatency.PacketID] = gobroLatency
	}

	latencies := make([]string, 0, len(gobroLatencyMap))

	for packetID := range gobroLatencyMap {
		if clientLatencyMap[packetID] == nil {
			continue
		}
		latencies = append(latencies, fmt.Sprint(float64(gobroLatencyMap[packetID].T.Sub(clientLatencyMap[packetID].T).Microseconds())/1000))
	}

	location := fmt.Sprint("data/latencyTests/", transportProtocol, "/client_", numberOfClients, "/")
	filename := fmt.Sprint(location, packetSize, "_bytes_rec.csv")
	if !serverReceive {
		filename = fmt.Sprint(location, packetSize, "_bytes_send.csv")
	}

	err := os.MkdirAll(location, os.ModePerm)
	if err != nil {
		panic(err)
	}

	csvFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	// Write num clients, server receive latency, server send latency
	for _, latency := range latencies {
		csvFile.Write([]byte(latency + ","))
	}
}
