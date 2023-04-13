package network

import (
	"MQTT-GO/structures"
	"net"
	"sync"
	"testing"
	"time"
)

func TestUDPnetwork(t *testing.T) {

	listener, _ := NewListener(UDP)
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

	connection, _ := NewCon(UDP)
	connection.Connect("localhost", 8000)
	connection.Write([]byte("djskaljda"))

	time.Sleep(time.Millisecond * 100)

	connection2, _ := NewCon(UDP)
	connection2.Connect("localhost", 8000)
	connection2.Write([]byte("1111"))
	waitGroup.Wait()

	buffer := make([]byte, 100)
	conn.Read(buffer)
	structures.Println(buffer)

	conn2.Read(buffer)
	structures.Println(buffer)

}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}
