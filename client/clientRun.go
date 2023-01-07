package client

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
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
	fmt.Println("Connecting to server on port", *port)

	ConnectToServer(*ip, *port)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		fmt.Println(text)

	}
}
