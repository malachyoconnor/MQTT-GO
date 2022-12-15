package client

import (
	"net"
	"sync"
)

type ClientID string

type ClientTable map[ClientID]*Client

type Client struct {
	// Should be a max of 23 characters!
	ClientIdentifier ClientID
	Topics           []string
	TCPConnection    net.Conn
}

var numClientsMutex sync.Mutex
var numClients int64 = 0

func generateClientID() ClientID {
	// TODO: Make sure this returns a better unique string per new client

	numClientsMutex.Lock()
	numClients += 1
	stringLen := numClients
	numClientsMutex.Unlock()

	var username string = ""

	for i := 0; i < int(stringLen); i++ {
		username += "a"
	}

	return ClientID(username)

}
