package stresstests

import (
	"MQTT-GO/client"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
	"flag"
	"os"
	"time"
)

var (
	doSubscribes = flag.Bool("continualSubscribe", false, "Continually subscribe")
	runServer    = flag.Bool("runServer", false, "Start a server")
)

const (
	hoursToTest = 3
)

func continualSubscribe(ip string, port int) {
	client := client.CreateClient()
	err := client.SetClientConnection(ip, port)
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
