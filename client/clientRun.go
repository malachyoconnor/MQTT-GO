package client

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"MQTT-GO/packets"
)

var (
	port = flag.Int("port", 8000, "Get the port the server is being run on")
	ip   = flag.String("ip", "localhost", "Get the ip the server is being run on")
)

// We want to listen on the command line for inputs
// E.g. Subscribe, publish, disconnect etc.
// Then send values to the server
func StartClient() {
	flag.Parse()

	client := CreateClient()
	listenForExit(client)
	err := client.SetClientConnection(*ip, *port)
	if err != nil {
		if err.Error()[len(err.Error())-len("connection refused"):] == "connection refused" {
			fmt.Println("Could not connect to server - connection was refused")
		} else {
			fmt.Println(err)
		}

		return
	}
	fmt.Println("Got here")
	err = client.SendConnect()
	fmt.Println("Got here")
	if err != nil {
		fmt.Println(err)
		return
	}

	// We need to start listening for responses AFTER we send the connect, because the connect packet doesn't have
	// a packet identifier to listen on.
	go client.ListenForPackets()

	fmt.Println("Connected to broker on port", *port)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		words := strings.Split(text, " ")

		if len(words) == 1 {
			continue
		}

		switch words[0] {
		case "publish":
			{
				var stringBuilder strings.Builder
				for _, word := range words[2:] {
					stringBuilder.WriteString(word)
					stringBuilder.WriteRune(' ')
				}

				err := client.SendPublish([]byte(stringBuilder.String()), words[1])
				if err != nil {
					log.Println("Error while sending subscribe", err)
				}
			}
		case "subscribe":
			{
				topics := make([]packets.TopicWithQoS, 0, len(words)-1)
				for _, word := range words[1:] {
					topics = append(topics, packets.TopicWithQoS{Topic: word})
				}
				err := client.SendSubscribe(topics...)
				if err != nil {
					log.Println("Error while sending subscribe", err)
				}
			}
		case "unsubscribe":
			{
				topicsToUnsub := words[1:]
				err := client.SendUnsubscribe(topicsToUnsub...)
				if err != nil {
					log.Println("Error while sending unsubscribe", err)
				}
			}
		}

	}
}
