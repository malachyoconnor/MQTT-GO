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

// ManyClientsPublish starts a number of clients, and publishes a message from each of them
// This is used to test the performance of the server
func ManyClientsPublish(numClients int, ip string, port int) {
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

	for _, openClient := range clients {
		go func(c client.Client) {
			for i := 0; i < 1000; i++ {
				err := c.SendPublish([]byte("TEST"), "abc")
				if err != nil {
					structures.Println("Error while publishing", err)
				}
			}
		}(openClient)
	}

	fmt.Fprintln(storedStdout, "PUBLISHED FROM ALL CLIENTS")

	for _, client := range clients {
		err := client.SendDisconnect()
		if err != nil {
			fmt.Fprintln(storedStdout, "Error while disconnecting", err)
		}
		time.Sleep(time.Millisecond)
		err = client.BrokerConnection.Close()
		if err != nil {
			fmt.Fprintln(storedStdout, "ERROR WHILE DISCONNECTING", err)
		}
	}

	fmt.Fprintln(storedStdout, "DISCONNECTED ALL THE CLIENTS")

}
