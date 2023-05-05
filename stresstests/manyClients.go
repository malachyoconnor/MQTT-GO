// Package stresstests contains functions that stress test the broker.
package stresstests

import (
	"MQTT-GO/client"
	"fmt"
	"os"
	"sync"
)

// ManyClientsConnect creates a large number of clients and connects them to the broker.
// It then sends a publish message from each client, and disconnects them.
func ManyClientsConnect(ip string, port int, numberOfClients int) {
	// Stop the clients from printing to stdout
	storedStdout := os.Stdout
	// os.Stdout = nil
	listenAndExit(storedStdout)

	go fmt.Fprintln(storedStdout, "\rNum clients:", numberOfClients)
	clients := make([]*client.Client, numberOfClients)
	go exitAll(clients)

	queue := sync.WaitGroup{}
	queue.Add(numberOfClients)

	connectAllClients(clients, ip, port, storedStdout)

	for _, client := range clients {
		err := client.SendPublish([]byte("TEST"), "abc")
		if err != nil {
			fmt.Println("Error while publishing", err)
		}
	}
	fmt.Fprintln(storedStdout, "PUBLISHED FROM ALL CLIENTS")

	disconnectAllClients(clients, storedStdout)
}
