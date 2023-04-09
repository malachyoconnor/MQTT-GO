package network

import (
	"fmt"
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
	fmt.Println(buffer)

	conn2.Read(buffer)
	fmt.Println(buffer)

}

func TestVarIntEncodeDecode(t *testing.T) {
	// I originally started testing with 1<<62
	// This would have taken 5 millenia to complete the test...
	for i := uint64(0); i < 1<<10; i++ {
		res, err := EncodeVarInt(i)
		checkErr(err, t)
		if x, offset, _ := DecodeVarInt(res); x != i || int(offset) != len(res) {
			fmt.Printf("start: %v encoded: %v decoded: %v encoded length: %v\n", i, res, x, offset)
			if int(offset) != len(res) {
				t.Error("Incorrect offset")
			}

			fmt.Printf("correct: %v. our value: %v - intermediate: %v", i, x, res)
			t.Error("Doesnt decode correctly")
		}
	}
}

func TestGetBits(t *testing.T) {
	tests := []struct {
		value byte
		bits  []byte
	}{
		{byte(127), []byte{1, 3, 4}},
		{byte(1), []byte{0, 3, 5}},
		{byte(63), []byte{0, 1, 2, 3, 4, 5, 6, 7}},
	}

	for _, test_case := range tests {
		fmt.Printf("%08b\n", getBits(test_case.value, test_case.bits...))
	}

}

func TestQUICVarLengthInt(t *testing.T) {

	for test_val := 0; test_val < 2048; test_val++ {

		encoded_val, err := EncodeVarInt(uint64(test_val))
		if err != nil {
			t.Error(err)
		}
		decoded_val, _, _ := DecodeVarInt(encoded_val)

		if decoded_val != uint64(test_val) {
			fmt.Println(decoded_val)
			t.Error(test_val, encoded_val, decoded_val)
		}
	}
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func TestTLS(t *testing.T) {
	// cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8000})
	// checkErr(err, t)
	// defer conn.Close()

	// buffer := make([]byte, 300)
	// _, err = conn.Read(buffer)
	// checkErr(err, t)

	// res, err := decodeLongHeaderPacket(buffer)
	// x := InitialPacket(*res.(*InitialPacket))
	// checkErr(err, t)

	// _, err = decodeFrame(x.PacketPayload)
	// checkErr(err, t)

	// fmt.Println()

	// structures.PrintInterface(x)
	// tconn := tls.Server(conn, config)

	// err = tconn.Handshake()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// tconn.Read(buffer)
	// tconn.Write([]byte{1, 2, 3})

	// fmt.Println(buffer)
}
