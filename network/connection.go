// Package network contains all the code for the network layer.
// This includes the connection and listener interfaces, as well as the implementations for TCP, UDP and QUIC.
package network

import (
	"fmt"
	"net"
	"sync"
	"time"

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

// Conn is an interface that allows us to switch between TCP, UDP and QUIC
// connections easily. For any type of transport protocol we want to support,
// we just need to implement this interface.
type Conn interface {
	Connect(ip string, port int) error
	Write(toWrite []byte) (n int, err error)
	Read(buffer []byte) (n int, err error)
	Close() error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// NewConn returns a new connection of the type specified by the networkID.
func NewConn(networkID byte) (Conn, error) {
	switch networkID {
	case TCP:
		{
			return &TCPConn{}, nil
		}
	case UDP:
		{
			return &UDPConn{}, nil
		}
	case QUIC:
		{
			return &QUICConn{
				streamReadLock:  &sync.Mutex{},
				streamWriteLock: &sync.Mutex{},
			}, nil
		}
	}
	return nil, fmt.Errorf("error: Supplied networkID %v is not defined", networkID)
}

// TCPConn is a struct that implements the Conn interface for TCP connections.
type TCPConn struct {
	connection *net.TCPConn
}

// UDPServerConnection is a constant that represents that the connection is the server.
// UDPClientConnection is a constant that represents that the connection is the client.
const (
	UDPServerConnection byte = 1
	UDPClientConnection byte = 2
)

// UDPConn is a struct that implements the Conn interface for UDP connections.
type UDPConn struct {
	connection     *net.UDPConn
	packetBuffer   chan []byte
	localAddr      net.Addr
	remoteAddr     net.Addr
	connected      bool
	connectionType byte

	serverConnectionDeleter func()
}

// QUICConn is a struct that implements the Conn interface for QUIC connections.
type QUICConn struct {
	connection      *quic.Connection
	stream          *quic.Stream
	streamReadLock  *sync.Mutex
	streamWriteLock *sync.Mutex
}
