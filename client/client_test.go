package client

import (
	"MQTT-GO/gobro"
	"fmt"
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

func TestMain(m *testing.M) {
	ServerUp()
	m.Run()
	server.StopServer()
}

func TestConnectToServer(t *testing.T) {
	client := CreateClient()
	fmt.Println("clientID", client.clientID)
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

	client := CreateClient()
	fmt.Println("clientID", client.clientID)
	err := client.SetClientConnection("localhost", 8000)
	if err != nil {
		t.Error(err)
	}
	err = client.SendConnect()
	defer client.SendDisconnect()

	if err != nil {
		t.Error("Error while connecting to server", err)
	}

	err = client.SendPublish([]byte("test"), "test")
	if err != nil {
		t.Error("Error while publishing to server:", err)
	}

}

func TestConstantPublish(t *testing.T) {
	ServerUp()
	client := CreateClient()
	fmt.Println("clientID", client.clientID)
	client.SetClientConnection("localhost", 8000)
	client.SendConnect()
	defer client.SendDisconnect()

	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		client.SendPublish([]byte(fmt.Sprint("test", i)), "x/y")
	}

}
