package client

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
)

type Client struct {
	clientID         string
	brokerConnection *net.Conn
	receivedMessages chan *[]byte
	waitingPackets   *waitingPackets
}

var (
	port = flag.Int("port", 8000, "Get the port the server is being run on")
	ip   = flag.String("ip", "localhost", "Get the ip the server is being run on")
)

// We want to listen on the command line for inputs
// E.g. Subscribe, publish, disconnect etc.
// Then send values to the server
func StartClient() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	client := CreateClient()
	err := client.SetClientConnection(*ip, *port)
	go client.ListenForPackets()
	if err != nil {
		if err.Error()[len(err.Error())-len("connection refused"):] == "connection refused" {
			fmt.Println("Could not connect to server - connection was refused")
		} else {
			fmt.Println(err)
		}

		return
	}
	err = client.SendConnect()
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for range c {
			client.SendDisconnect()
			(*client.brokerConnection).Close()
			fmt.Println("\nConnection closed, goodbye")
			os.Exit(0)
		}
	}()

	fmt.Println("Connected to broker on port", *port)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		words := strings.Split(text, " ")

		switch words[0] {
		case "publish":
			{
				var stringBuilder strings.Builder
				if len(words) == 1 {
					continue
				}
				for _, word := range words[2:] {
					stringBuilder.WriteString(word)
					stringBuilder.WriteRune(' ')
				}

				client.SendPublish([]byte(stringBuilder.String())[:], words[1])
			}
		case "subscribe":
			{
				client.SendSubscribe(topicWithQoS{topic: "DoesThisWork"})

			}
		}

	}
}

func CreateClient() *Client {

	messageChannel := make(chan *[]byte, 20)
	waitingPackets := CreateWaitingPacketList()
	return &Client{
		receivedMessages: messageChannel,
		clientID:         generateRandomClientID(),
		waitingPackets:   waitingPackets,
	}
}

func (client *Client) SetClientConnection(ip string, port int) error {
	connection, err := net.Dial("tcp", net.JoinHostPort(ip, fmt.Sprint(port)))
	if err != nil {
		return err
	}

	client.brokerConnection = &connection
	return nil
}
