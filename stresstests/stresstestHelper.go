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

var numPublished atomic.Int32

func connectAllClients(clientList []*client.Client, ip string, port int, storedStdout *os.File) {
	queue := sync.WaitGroup{}
	queue.Add(len(clientList))

	for i := 0; i < len(clientList); i++ {
		go func(clientNum int) {
			newClient, err := client.CreateAndConnectClient(ip, port)
			if err != nil {
				fmt.Fprintln(storedStdout, "Error while connecting:", err)
				return
			}

			clientList[clientNum] = newClient
			fmt.Fprint(storedStdout, "\rClients created and connected:", numPublished.Add(1))
			if err != nil {
				fmt.Fprintln(storedStdout, "Error during publish", err)
			}
			queue.Done()
		}(i)

		time.Sleep(time.Millisecond * 5)
	}
	queue.Wait()
	fmt.Fprintln(storedStdout, "CONNECTED ALL CLIENTS")
}

func disconnectAllClients(clientList []*client.Client, storedStdout *os.File) {
	for _, client := range clientList {
		if client.BrokerConnection == nil {
			fmt.Println("error: NIL CONNECTION")
			continue
		}
		err := client.SendDisconnect()
		if err != nil {
			fmt.Println(storedStdout, "Error while disconnecting", err)
		}
		time.Sleep(5 * time.Millisecond)
		err = client.BrokerConnection.Close()
		if err != nil {
			fmt.Println(storedStdout, "ERROR WHILE DISCONNECTING", err)
		}
	}

	fmt.Println("DISCONNECTED ALL THE CLIENTS")
}

func exitAll(clients []*client.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		fmt.Println("Exiting forcibly")
		queue := sync.WaitGroup{}
		queue.Add(len(clients))
		for _, openClient := range clients {
			if openClient.BrokerConnection != nil {
				go func(c *client.Client) {
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

func listenAndExit(storedStdout *os.File) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for range c {
			fmt.Fprintln(storedStdout, "Interrupted")
			os.Exit(1)
		}
	}()

}
