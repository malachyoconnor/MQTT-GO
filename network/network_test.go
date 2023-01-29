package network_test

import (
	"MQTT-GO/network"
	"fmt"
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

	connection, _ := network.NewCon(network.UDP)
	connection.Connect("localhost", 8000)
	connection.Write([]byte("djskaljda"))

	time.Sleep(time.Millisecond * 100)

	connection2, _ := network.NewCon(network.UDP)
	connection2.Connect("localhost", 8000)
	connection2.Write([]byte("1111"))
	waitGroup.Wait()

	buffer := make([]byte, 100)
	conn.Read(buffer)
	fmt.Println(buffer)

	conn2.Read(buffer)
	fmt.Println(buffer)

}

func TestVarIntEncodeDecode(t *testing.T) {

	for i := uint64(0); i < 1<<62; i++ {
		res, _ := network.EncodeVarInt(i)

		if x := network.DecodeVarInt(res); x != i {

			fmt.Printf("correct: %v. our value: %v", i, x)

			t.Errorf("Doesnt decode correctly")
		}

	}

}
