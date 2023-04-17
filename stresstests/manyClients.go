package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/structures"
	"fmt"
	"os"
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

	fmt.Fprintln(StoredStdout, "\rNum clients:", numClients)

	clients := make([]client.Client, numClients)

	queue := sync.WaitGroup{}
	queue.Add(numClients)

	var numPublished atomic.Int32

	for i := 0; i < numClients; i++ {

		go func(clientNum int) {
			newClient, err := client.CreateAndConnectClient("localhost", 8000)
			for err != nil {
				structures.Println(err)
				newClient, err = client.CreateAndConnectClient("localhost", 8000)
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
	fmt.Fprintln(StoredStdout, "CONNECTED ALL CLIENTS")

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
