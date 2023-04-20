package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/structures"
	"fmt"
	"os"
)

// ManyClientsPublish starts a number of clients, and publishes a message from each of them
// This is used to test the performance of the server
func ManyClientsPublish(numClients int, ip string, port int) {
	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	os.Stdout = nil

	go fmt.Fprintln(storedStdout, "\rNum clients:", numClients)
	listenAndExit(storedStdout)

	clients := make([]client.Client, numClients)
	go exitAll(clients)
	connectAllClients(clients, ip, port, storedStdout)

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

	disconnectAllClients(clients, storedStdout)
}
