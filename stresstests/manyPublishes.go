package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/network"
	"MQTT-GO/packets"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	numberOfPublishes = 100
)

// ManyClientsPublish starts a number of clients, and publishes a message from each of them
// This is used to test the performance of the server
func ManyClientsPublish(ip string, port int, messageSize int, numberOfClients int) (packetsReceived int, expectedPackets int) {
	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	// os.Stdout = nil
	fmt.Print("Contacting", ip, ":", port, "\n")

	go fmt.Fprintln(storedStdout, "\rNum clients:", numberOfClients)
	listenAndExit(storedStdout)

	clients := make([]*client.Client, numberOfClients)

	go exitAll(clients)
	connectAllClients(clients, ip, port, storedStdout)

	toPublish := make([]byte, messageSize)

	queue := sync.WaitGroup{}
	firstSubscriber := clients[0]
	firstSubscriber.SendPublish([]byte("Test"), "abc")
	err := firstSubscriber.SendSubscribe(packets.TopicWithQoS{Topic: "abc", QoS: 0})
	if err != nil {
		panic(err)
	}

	fmt.Println("RUNNING CLIENTS")

	for _, openClient := range clients[1:] {
		queue.Add(1)
		go func(c *client.Client) {
			for i := 0; i < numberOfPublishes; i++ {
				err := c.SendPublish(toPublish, "abc")
				if err != nil {
					fmt.Println("Error while publishing", err)
					panic(err)
				}
				time.Sleep(1 * time.Millisecond)
			}
			queue.Done()
		}(openClient)
	}
	queue.Wait()

	counter := 0
	for firstSubscriber.ReceivedPackets.Size() < (numberOfClients-1)*numberOfPublishes {
		time.Sleep(100 * time.Millisecond)
		counter++
		if ConnectionType == network.UDP {
			counter += 9
		}
		if counter > 100 {
			break
		}
		fmt.Println("Waiting for all packets to be received", firstSubscriber.ReceivedPackets.Size(), (numberOfClients-1)*numberOfPublishes)
	}

	fmt.Println("PUBLISHED FROM ALL CLIENTS")

	disconnectAllClients(clients, storedStdout)
	return firstSubscriber.ReceivedPackets.Size(), (numberOfClients - 1) * numberOfPublishes
}
