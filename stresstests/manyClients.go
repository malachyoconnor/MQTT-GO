// Package stresstests contains functions that stress test the broker.
package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/structures"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

// ManyClientsConnect creates a large number of clients and connects them to the broker.
// It then sends a publish message from each client, and disconnects them.
func ManyClientsConnect(numClients int, ip string, port int) {
	if numClients <= 0 {
		numClients = 100
	}
	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	os.Stdout = nil

	go fmt.Fprintln(storedStdout, "\rNum clients:", numClients)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for range c {
			fmt.Fprintln(storedStdout, "Interrupted")
			os.Exit(1)
		}
	}()

	clients := make([]client.Client, numClients)
	go exitAll(clients)

	queue := sync.WaitGroup{}
	queue.Add(numClients)

	var numPublished atomic.Int32

	for i := 0; i < numClients; i++ {
		go func(clientNum int) {
			newClient, err := client.CreateAndConnectClient(ip, port)
			if err != nil {
				fmt.Fprintln(storedStdout, "Error while connecting:", err)
				return
			}

			clients[clientNum] = *newClient
			err = newClient.SendPublish([]byte("TEST"), "abc")
			fmt.Fprint(storedStdout, "\rClients created and published:", numPublished.Add(1))
			if err != nil {
				fmt.Fprintln(storedStdout, "Error during publish", err)
			}
			queue.Done()
		}(i)

		time.Sleep(time.Millisecond * 5)
	}
	queue.Wait()
	fmt.Fprintln(storedStdout, "\nCONNECTED ALL CLIENTS")

	for _, client := range clients {
		err := client.SendPublish([]byte("TEST"), "abc")
		if err != nil {
			structures.Println("Error while publishing", err)
		}
	}

	fmt.Fprintln(storedStdout, "PUBLISHED FROM ALL CLIENTS")

	for _, openClient := range clients {
		go func(client client.Client) {
			err := client.SendDisconnect()
			if err != nil {
				fmt.Fprintln(storedStdout, "Error while disconnecting", err)
			}
			time.Sleep(time.Millisecond)
			err = client.BrokerConnection.Close()
			if err != nil {
				fmt.Fprintln(storedStdout, "ERROR WHILE DISCONNECTING", err)
			}
		}(openClient)
	}

	fmt.Fprintln(storedStdout, "DISCONNECTED ALL THE CLIENTS")

}

func exitAll(clients []client.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		queue := sync.WaitGroup{}
		queue.Add(len(clients))
		for _, openClient := range clients {
			if openClient.BrokerConnection != nil {
				go func(c client.Client) {
					err := c.SendDisconnect()
					if err != nil {
						structures.PrintCentrally("Error while disconnecting", err)
						return
					}
					time.Sleep(time.Millisecond * 3)
					err = c.BrokerConnection.Close()
					if err != nil {
						structures.PrintCentrally("Error while closing", err)
					}
					queue.Done()
				}(openClient)
			}
		}
		queue.Wait()
		os.Exit(0)
	}
}
