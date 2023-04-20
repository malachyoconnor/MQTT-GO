package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

// Listener is an interface that allows us to switch between TCP, UDP and QUIC
// listeners easily. For any type of transport protocol we want to support,
// we just need to implement this interface.
// I.e. implement Listen, Accept and Close
type Listener interface {
	Listen(ip string, port int) error
	Accept() (net.Conn, error)
	Close() error
}

// TCPListener is a struct that implements the Listener interface for TCP listeners.
type TCPListener struct {
	listener *net.Listener
}

// NewListener returns a new listener of the type specified by the networkID.
func NewListener(networkID byte) (Listener, error) {
	switch networkID {
	case TCP:
		{
			return &TCPListener{}, nil
		}
	case UDP:
		{
			return &UDPListener{}, nil
		}
	case QUIC:
		{
			return &QUICListener{}, nil
		}
	}
	return nil, fmt.Errorf("error: Supplied networkID %v is not defined", networkID)
}

// UDPListener is a struct that implements the Listener interface for UDP listeners.
type UDPListener struct {
	listener *net.UDPConn
	// The string is the address:port as you would expect
	openConnections     map[string]chan []byte
	openConnectionsLock sync.RWMutex
	listening           bool
	// If Accept() is called, then the listener pushes the new clients to a newClientBuffer
	// these then get picked up by Accept.
	newClientBuffer chan string
	localAddr       *net.UDPAddr
}

// QUICListener is a struct that implements the Listener interface for QUIC listeners.
type QUICListener struct {
	listener *quic.Listener
}
