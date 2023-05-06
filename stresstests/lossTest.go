package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/packets"
	"fmt"
	"os"
	"time"
)

func TestUDPLoss(numPackets int, ip string, port int, packetSize int, numberOfClients int) {
	fmt.Println("Testing UDP Loss")

	transportProtocol := []string{"TCP", "QUIC", "UDP"}[ConnectionType]
	fmt.Println("Transport protocol:", transportProtocol)

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
		// UDP will be more lossy, so we don't need to wait as long
		if transportProtocol == "UDP" {
			counter += 9
		}
		if counter > 100 {
			break
		}
		fmt.Println("Waiting for all packets to be received", newClient.ReceivedPackets.Size(), numPackets)
	}

	// Write num clients, packets received, expected packets to csv
	location := fmt.Sprint("data/lossTests/", transportProtocol, "/packet_", packetSize, "B/")
	filename := fmt.Sprint(location, numberOfClients, "_clients.csv")

	err = os.MkdirAll(location, os.ModePerm)
	if err != nil {
		panic(err)
	}

	csvFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	_, err = csvFile.Write([]byte(fmt.Sprint(packetSize, ",", newClient.ReceivedPackets.Size(), ",", numPackets, ",")))
	if err != nil {
		panic(err)
	}

}
