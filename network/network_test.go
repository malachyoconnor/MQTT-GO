package network_test

import (
	"MQTT-GO/network"
	"MQTT-GO/structures"
	"net"
	"sync"
	"testing"
	"time"
)

func TestUDPnetwork(t *testing.T) {

	listener, _ := network.NewListener(network.UDP)
	err := listener.Listen("localhost", 8000)
	if err != nil {
		t.Error(err)
	}
	var (
		conn  net.Conn
		conn2 net.Conn
	)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	go func() {
		conn, _ = listener.Accept()
		conn2, _ = listener.Accept()
		waitGroup.Done()
	}()

	connection, _ := network.NewConn(network.UDP)
	connection.Connect("localhost", 8000)
	connection.Write([]byte("djskaljda"))

	time.Sleep(time.Millisecond * 100)

	connection2, _ := network.NewConn(network.UDP)
	connection2.Connect("localhost", 8000)
	connection2.Write([]byte("1111"))
	waitGroup.Wait()

	buffer := make([]byte, 100)
	conn.Read(buffer)
	structures.Println(buffer)

	conn2.Read(buffer)
	structures.Println(buffer)

}
