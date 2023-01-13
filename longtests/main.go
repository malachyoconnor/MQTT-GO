package main

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/packets"
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	doSubscribes = flag.Bool("continualSubscribe", false, "Continually subscribe")
	runServer    = flag.Bool("runServer", false, "Start a server")
)

func main() {
	fmt.Println("Starting")
	flag.Parse()
	switch true {
	case *doSubscribes:
		continualSubscribe()
	case *runServer:
		server := gobro.CreateServer()
		server.StartServer()
	}
}

const (
	hoursToTest = 3
)

func continualSubscribe() {
	client := client.CreateClient()
	err := client.SetClientConnection("localhost", 8000)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.SendConnect()
	go client.ListenForPackets()

	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.SendSubscribe(packets.TopicWithQoS{Topic: "continual"})
	if err != nil {
		fmt.Println(err)
		return
	}

	message := []byte("This is my test message")
	fmt.Println("Starting the test")
	go func() {
		time.Sleep((3600*hoursToTest + 100) * time.Second)
		fmt.Println("We waited too long!! Exiting")
		os.Exit(0)
	}()

	for i := 0; i < 3600*hoursToTest; i++ {
		time.Sleep(time.Second)
		err := client.SendPublish(append(message, byte(i)), "continual")
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Completed successfully")

}
