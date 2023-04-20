package client

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

// StartClient starts the client, and listens for inputs from the command line.
// E.g. Subscribe, publish, disconnect etc, Then communicates with the broker.
func StartClient(ip string, port int) {
	flag.Parse()
	client := CreateClient()
	listenForExit(client)
	err := client.SetClientConnection(ip, port)

	if err != nil {
		structures.Println(err)
		return
	}

	err = client.SendConnect(ip, port)
	if err != nil {
		structures.Println(err)
		return
	}

	// We need to start listening for responses AFTER we send the connect, because the connect packet doesn't have
	// a packet identifier to listen on.
	go client.ListenForPackets()

	structures.Println("Connected to broker on port", port)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// Windows uses \r\n - we need to get them out of there
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.ReplaceAll(text, "\n", "")
		words := strings.Split(text, " ")
		if len(words) == 1 {
			continue
		}

		for i, word := range words {
			words[i] = strings.ReplaceAll(word, "\r", "")
		}

		var err error

		switch words[0] {
		case "publish":
			{
				combinedWords := strings.Join(words[2:], " ")

				err = client.SendPublish([]byte(combinedWords), words[1])
			}
		case "subscribe":
			{
				topics := make([]packets.TopicWithQoS, 0, len(words)-1)
				for _, word := range words[1:] {
					topics = append(topics, packets.TopicWithQoS{Topic: word})
				}
				err = client.SendSubscribe(topics...)
			}
		case "unsubscribe":
			{
				topicsToUnsub := words[1:]
				err = client.SendUnsubscribe(topicsToUnsub...)
			}
		}

		if err != nil {
			log.Println("Error while sending", words[0], err)
		}
	}
}
