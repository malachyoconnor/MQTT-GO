package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

type Listener interface {
	Listen(ip string, port int) error
	Accept() (net.Conn, error)
	Close() error
}

type TCPListener struct {
	listener *net.Listener
}

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

type QUICListener struct {
	listener *quic.Listener
}
