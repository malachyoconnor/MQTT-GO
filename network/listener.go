package network

import (
	"fmt"
	"net"
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
	}
	return nil, fmt.Errorf("error: Supplied networkID %v is not defined", networkID)
}
