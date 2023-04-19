package main

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"flag"
	"os"
	"time"
)

var (
	doSubscribes = flag.Bool("continualSubscribe", false, "Continually subscribe")
	runServer    = flag.Bool("runServer", false, "Start a server")
	ip           = flag.String("testip", "127.0.0.1", "Ip")
	port         = flag.Int("testport", 8000, "Port")
)

func main() {
	structures.Println("Starting")
	flag.Parse()
	switch {
	case *doSubscribes:
		continualSubscribe()
	case *runServer:
		server := gobro.NewServer()
		server.StartServer("localhost", 8000)
	}
}

const (
	hoursToTest = 3
)

func continualSubscribe() {
	client := client.CreateClient()
	err := client.SetClientConnection(*ip, *port)
	if err != nil {
		structures.Println(err)
		return
	}

	err = client.SendConnect("localhost", 8000)
	go client.ListenForPackets()

	if err != nil {
		structures.Println(err)
		return
	}

	err = client.SendSubscribe(packets.TopicWithQoS{Topic: "continual"})
	if err != nil {
		structures.Println(err)
		return
	}

	message := []byte("This is my test message")
	structures.Println("Starting the test")
	go func() {
		time.Sleep((3600*hoursToTest + 100) * time.Second)
		structures.Println("We waited too long!! Exiting")
		os.Exit(0)
	}()

	for i := 0; i < 3600*hoursToTest; i++ {
		time.Sleep(time.Second)
		err := client.SendPublish(append(message, byte(i)), "continual")
		if err != nil {
			structures.Println(err)
		}
	}

	structures.Println("Completed successfully")

}
