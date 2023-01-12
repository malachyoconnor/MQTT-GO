package client_test

import (
	"MQTT-GO/client"
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

func testErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestConnectToServer(t *testing.T) {
	client := client.CreateClient()
	fmt.Println("clientID", client.ClientID)
	defer testErr(t, client.SendDisconnect())
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

	client := client.CreateClient()
	fmt.Println("clientID", client.ClientID)
	err := client.SetClientConnection("localhost", 8000)
	if err != nil {
		t.Error(err)
	}
	err = client.SendConnect()
	defer testErr(t, client.SendDisconnect())

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
	client := client.CreateClient()
	fmt.Println("clientID", client.ClientID)
	err := client.SetClientConnection("localhost", 8000)
	testErr(t, err)
	defer testErr(t, client.SendDisconnect())

	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		err := client.SendPublish([]byte(fmt.Sprint("test", i)), "x/y")
		testErr(t, err)
	}

}

func TestWaitingPackets(t *testing.T) {

	waitingPacketsList := client.CreateWaitingPacketList()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		waitingPacketsList.GetOrWait(0)
		waitGroup.Add(-1)
	}()
	go func() {
		waitingPacketsList.GetOrWait(2)
		waitGroup.Add(-1)
	}()

	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 1})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 5})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 9})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 2})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 0})
	waitGroup.Wait()
}
