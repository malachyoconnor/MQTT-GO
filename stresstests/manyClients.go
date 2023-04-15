package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/structures"
	"fmt"
	"os"
)

func ConnectAndPublish(numClients int) {
	if numClients <= 0 {
		numClients = 100
	}
	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	os.Stdout = nil

	fmt.Fprintln(storedStdout, "\rNum clients:", numClients)

	clients := make([]client.Client, numClients)

	for i := 0; i < numClients; i++ {
		fmt.Fprint(storedStdout, "\rClients created and connected:", i+1)
		client, err := client.CreateAndConnectClient("localhost", 8000)
		structures.PANIC_ON_ERR(err)

		clients[i] = *client

		err = client.SendPublish([]byte("TEST"), "abc")
		structures.PANIC_ON_ERR(err)
	}

	fmt.Fprintln(storedStdout, "CONNECTED ALL CLIENTS")

	for _, client := range clients {
		client.SendPublish([]byte("TEST"), "abc")
	}

	fmt.Fprintln(storedStdout, "PUBLISHED FROM ALL CLIENTS")

	for _, client := range clients {
		err := client.SendDisconnect()
		if err != nil {
			fmt.Fprintln(storedStdout, "ERROR WHILE DISCONNECTING", err)
		}
	}

	fmt.Fprintln(storedStdout, "DISCONNECTED ALL THE CLIENTS")

}
