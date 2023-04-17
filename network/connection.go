package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
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
	LocalAddr() net.Addr
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

const (
	QUIC_SERVER_CONNECTION byte = 1
	QUIC_CLIENT_CONNECTIOn byte = 2
)

type QUICCon struct {
	connection      *quic.Connection
	stream          *quic.Stream
	streamReadLock  *sync.Mutex
	streamWriteLock *sync.Mutex
}
