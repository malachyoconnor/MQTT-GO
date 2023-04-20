// Package network contains all the code for the network layer.
// This includes the connection and listener interfaces, as well as the implementations for TCP, UDP and QUIC.
package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

// This is a list of all the transport types we support and their IDs.
const (
	TCP  byte = 0
	QUIC byte = 1
	UDP  byte = 2
)

// We want to be able to switch easily between sending via TCP, UDP and QUIC.
// We want to be able to use the same functions for all three.

// Con is an interface that allows us to switch between TCP, UDP and QUIC
// connections easily. For any type of transport protocol we want to support,
// we just need to implement this interface.
type Con interface {
	Connect(ip string, port int) error
	Write(toWrite []byte) (n int, err error)
	Read(buffer []byte) (n int, err error)
	Close() error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}

// NewCon returns a new connection of the type specified by the networkID.
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
	case QUIC:
		{
			return &QUICCon{
				streamReadLock:  &sync.Mutex{},
				streamWriteLock: &sync.Mutex{},
			}, nil
		}
	}
	return nil, fmt.Errorf("error: Supplied networkID %v is not defined", networkID)
}

// TCPCon is a struct that implements the Con interface for TCP connections.
type TCPCon struct {
	connection *net.Conn
}

// UDPServerConnection is a constant that represents that the connection is the server.
// UDPClientConnection is a constant that represents that the connection is the client.
const (
	UDPServerConnection byte = 1
	UDPClientConnection byte = 2
)

// UDPCon is a struct that implements the Con interface for UDP connections.
type UDPCon struct {
	connection     *net.UDPConn
	packetBuffer   chan []byte
	localAddr      string
	remoteAddr     string
	connected      bool
	connectionType byte

	serverConnectionDeleter func()
}

// QUICCon is a struct that implements the Con interface for QUIC connections.
type QUICCon struct {
	connection      *quic.Connection
	stream          *quic.Stream
	streamReadLock  *sync.Mutex
	streamWriteLock *sync.Mutex
}
