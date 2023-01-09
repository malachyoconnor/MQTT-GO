package client

import (
	"MQTT-GO/gobro"
	"sync"
	"testing"
	"time"
)

var (
	server       gobro.Server
	serverUpLock sync.Mutex = sync.Mutex{}
	serverUp     bool       = false
)

func ServerUp() {
	serverUpLock.Lock()
	defer serverUpLock.Unlock()
	if serverUp {
		return
	}
	server = gobro.CreateServer()

	go func() {
		server.StartServer()
	}()
	serverUp = true
	time.Sleep(time.Millisecond * 10)
}

func TestConnectToServer(t *testing.T) {
	ServerUp()

	client := Client{}
	defer client.SendDisconnect()
	err := client.SetClientConnection("localhost", 8000)

	if err != nil {
		t.Error("Error while getting TCP connection to server", err)
	}

	err = client.SendConnect()

	if err != nil {
		t.Error("Error while sending CONNECT to server:", err)
	}

}

func TestPublishToServer(t *testing.T) {
	ServerUp()

	client := Client{}
	client.SetClientConnection("localhost", 8000)
	err := client.SendConnect()
	defer client.SendDisconnect()

	if err != nil {
		t.Error("Error while connecting to server", err)
	}

	err = client.SendPublish([]byte("test"), "test")
	if err != nil {
		t.Error("Error while publishing to server:", err)
	}

}
