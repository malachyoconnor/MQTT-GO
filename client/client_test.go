package client

import (
	"MQTT-GO/gobro"
	"fmt"
	"sync"
	"testing"
)

var (
	serverUp    bool
	serverMutex sync.Mutex
	server      gobro.Server
)

func ServerUp() {
	serverMutex.Lock()

	go func() {

		server = gobro.Server{}
		server.StartServer()

	}()

	serverMutex.Unlock()
}

func TestConnectToServer(t *testing.T) {
	ServerUp()
	// server.StopServer()

	err := ConnectToServer("localhost", 8000)

	if err != nil {
		fmt.Println(err)
	}

}
