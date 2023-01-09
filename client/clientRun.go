package client

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type Client struct {
	brokerConnection *net.Conn
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

	client := Client{}
	client.SetClientConnection(*ip, *port)
	client.SendConnect()

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
				client.SendPublish(&[]byte{'u', 'm', 'm', ' ', '?'}, "x/y")
			}
		}

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
