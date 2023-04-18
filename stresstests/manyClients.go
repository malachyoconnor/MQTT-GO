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

func ConnectAndPublish(numClients int) {
	if numClients <= 0 {
		numClients = 100
	}
	// Stop the clients from printing to stdout
	StoredStdout := os.Stdout
	os.Stdout = nil

	go fmt.Fprintln(StoredStdout, "\rNum clients:", numClients)

	clients := make([]client.Client, numClients)
	go exitAll(clients)

	queue := sync.WaitGroup{}
	queue.Add(numClients)

	var numPublished atomic.Int32

	for i := 0; i < numClients; i++ {

		go func(clientNum int) {
			newClient, err := client.CreateAndConnectClient("127.0.0.1", 8000)
			for err != nil {
				fmt.Fprintln(StoredStdout, "Error while connecting:", err)
				newClient, err = client.CreateAndConnectClient("127.0.0.1", 8000)
			}

			clients[clientNum] = *newClient
			err = newClient.SendPublish([]byte("TEST"), "abc")
			fmt.Fprint(StoredStdout, "\rClients created and published:", numPublished.Add(1))
			structures.PANIC_ON_ERR(err)
			// fmt.Fprint(storedStdout, "\rPublished:", numPublished.Load())
			queue.Done()
		}(i)

		time.Sleep(time.Millisecond * 5)
	}
	queue.Wait()
	fmt.Fprintln(StoredStdout, "\nCONNECTED ALL CLIENTS")

	for _, client := range clients {
		err := client.SendPublish([]byte("TEST"), "abc")
		if err != nil {
			structures.Println("Error while publishing", err)
		}
	}

	fmt.Fprintln(StoredStdout, "PUBLISHED FROM ALL CLIENTS")

	for _, client := range clients {
		err := client.SendDisconnect()
		if err != nil {
			fmt.Fprintln(StoredStdout, "Error while disconnecting", err)
		}
		time.Sleep(time.Millisecond)
		err = client.BrokerConnection.Close()
		if err != nil {
			fmt.Fprintln(StoredStdout, "ERROR WHILE DISCONNECTING", err)
		}
	}

	fmt.Fprintln(StoredStdout, "DISCONNECTED ALL THE CLIENTS")

}

func exitAll(clients []client.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		queue := sync.WaitGroup{}
		queue.Add(len(clients))
		for _, c := range clients {
			if c.BrokerConnection != nil {

				go func(c client.Client) {
					c.SendDisconnect()
					time.Sleep(time.Millisecond * 3)
					err := c.BrokerConnection.Close()
					if err != nil {
						structures.PrintCentrally("Error while closing", err)
					}
					queue.Done()
				}(c)
			}
		}
		queue.Wait()
		os.Exit(0)
	}
}
