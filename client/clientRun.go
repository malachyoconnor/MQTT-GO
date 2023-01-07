package client

import (
	"flag"
	"fmt"
	"os"
)

var port = flag.Int("port", 8000, "Get the port the server is being run on")

// We want to listen on the command line for inputs
// E.g. Subscribe, publish, disconnect etc.
// Then send values to the server
func StartClient() {
	flag.Parse()

	args := os.Args[1:]
	fmt.Println(args)
	fmt.Println(*port)

}
