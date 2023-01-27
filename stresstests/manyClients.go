package stresstests

import (
	"MQTT-GO/client"
	"fmt"
	"os"
)

func ConnectAndPublish(numClients int) {
	if numClients == 0 {
		numClients = 100
	}

	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	os.Stdout = nil

	clients := make([]client.Client, numClients)

	for i := 0; i < numClients; i++ {
		client, err := client.CreateAndConnectClient("localhost", 8000)
		if err != nil {
			fmt.Fprint(storedStdout, "Error: ", err)
		}
		clients[i] = *client

		client.SendPublish([]byte("dsadsadsadsa"), "abc")
	}

	fmt.Fprintln(storedStdout, "Connected all the clients")

	for _, client := range clients {
		client.SendPublish([]byte("dsadsadsadsa"), "abc")
	}

	fmt.Fprintln(storedStdout, "Published from all clients")

	for _, client := range clients {
		err := client.SendDisconnect()
		if err != nil {
			fmt.Fprintln(storedStdout, "ERROR while disconnecting", err)
		}
	}

	fmt.Fprintln(storedStdout, "Disconnected all the clients")

}
