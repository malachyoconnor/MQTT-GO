package client_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/packets"
	"MQTT-GO/structures"
)

var (
	server       gobro.Server
	serverUpLock = sync.Mutex{}
	serverUp     = false
)

func ServerUp() {
	serverUpLock.Lock()
	defer serverUpLock.Unlock()
	if serverUp {
		return
	}
	server = gobro.NewServer()

	go func() {
		server.StartServer("localhost", 8000)
	}()
	serverUp = true
	time.Sleep(time.Millisecond * 100)
}

func TestMain(m *testing.M) {
	fmt.Println("Starting server")
	ServerUp()
	gobro.DisableStdOutput()

	time.Sleep(time.Millisecond * 500)
	m.Run()

	// os.Exit(exitCode)
}

func testErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestConnectToServer(t *testing.T) {
	client, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)

	defer client.SendDisconnect()

	if err != nil {
		t.Error("Error while sending CONNECT to server:", err)
	}
}

func TestPublishToServer(t *testing.T) {
	client, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)
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
	client, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)
	defer client.SendDisconnect()

	for i := 0; i < 10; i++ {
		err := client.SendPublish([]byte(fmt.Sprint("test", i)), "x/y")
		testErr(t, err)
	}
	fmt.Println("Done publishing")
}

func TestWaitingPackets(t *testing.T) {
	waitingPacketsList := client.CreateWaitingAckList()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		waitingPacketsList.GetOrWait(0)
		waitGroup.Done()
	}()
	go func() {
		waitingPacketsList.GetOrWait(2)
		waitGroup.Done()
	}()

	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 1})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 5})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 9})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 2})
	waitingPacketsList.AddItem(&client.StoredPacket{PacketID: 0})
	waitGroup.Wait()
}

func TestReceivingPublish(t *testing.T) {
	client1, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)
	client2, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)

	client1.SendSubscribe(packets.TopicWithQoS{Topic: "testing"})
	client2.SendPublish([]byte("test message"), "testing")

	done := false

	go func() {
		time.Sleep(time.Millisecond * 50)
		if !done {
			t.Error("publish took too long")
		}
	}()

	structures.PrintInterface(client1.ReceivedPackets.Head())
	done = true
}

func TestReceivingMultiplePublishes(t *testing.T) {
	client1, err := client.CreateAndConnectClient("localhost", 8000)
	testErr(t, err)
	err = client1.SendSubscribe(packets.TopicWithQoS{Topic: "testing"})
	testErr(t, err)
	publishClients := make([]*client.Client, 10)

	for i := 0; i < 10; i++ {
		publishClients[i], err = client.CreateAndConnectClient("localhost", 8000)
		testErr(t, err)
	}

	for i, client := range publishClients {
		err = client.SendPublish([]byte(fmt.Sprintln("Testing", i)), "testing")
		testErr(t, err)
	}

	done := false

	go func() {
		time.Sleep(time.Millisecond * 500)
		if !done {
			t.Error("publish took too long")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	if client1.ReceivedPackets.Size() != 10 {
		t.Error("Received packets size is not 10")
	}

	for i := 0; i < 10; i++ {
		client1.ReceivedPackets.GetItems()
		structures.Println("Got", i)
	}
	done = true
}
