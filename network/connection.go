package network

import (
	"fmt"
	"net"
)

const (
	TCP  byte = 0
	QUIC byte = 1
	UDP  byte = 2
)

// We want to be able to switch easily between sending via TCP, UDP and QUIC.
// We want to be able to use the same functions for all three.

type Con interface {
	Connect(ip string, port int) error
	Write(toWrite []byte) (n int, err error)
	Read(buffer []byte) (n int, err error)
	Close() error
	RemoteAddr() net.Addr
}

func NewCon(networkID byte) (Con, error) {
	switch networkID {
	case TCP:
		{
			return &TCPCon{}, nil
		}
	case UDP:
		{
			return &UDPCon{}, nil
		}
	}
	return nil, fmt.Errorf("error: Supplied networkID %v is not defined", networkID)
}

type TCPCon struct {
	connection *net.Conn
}

const (
	UDP_SERVER_CONNECTION byte = 1
	UDP_CLIENT_CONNECTION byte = 2
)

type UDPCon struct {
	connection     *net.UDPConn
	packetBuffer   chan []byte
	localAddr      string
	remoteAddr     string
	connected      bool
	connectionType byte

	serverConnectionDeleter func()
}
