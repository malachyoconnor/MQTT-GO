package network

import (
	"fmt"
	"net"

	"github.com/goburrow/quic"
	"github.com/goburrow/quic/transport"
)

type quicClientHandler struct {
	toWrite    chan []byte
	mainStream *quic.Stream
	client     *quic.Client
}

func (handler *quicClientHandler) Serve(conn *quic.Conn, events []transport.Event) {

	if len(handler.toWrite) != 0 {
		conn.StreamWrite(0, <-handler.toWrite)
	}

	for _, e := range events {

		switch e.Type {
		case transport.EventConnOpen:
			{
				st, err := conn.Stream(e.Data)
				if err != nil {
					panic(err)
				}
				handler.mainStream = st
			}
		}
	}
}

type quicServerHandler struct {
	waitingConnections  chan *quic.Stream
	acceptedConnections []*quic.Conn
	connectionAccepted  map[net.Addr]bool
	quicServer          *quic.Server
}

func (handler *quicServerHandler) Serve(conn *quic.Conn, events []transport.Event) {

	for _, e := range events {
		switch e.Type {

		case transport.EventStreamOpen:
			{
				st, err := conn.Stream(e.Data)
				if err != nil {
					fmt.Println(err)
					continue
				}
				handler.waitingConnections <- st
				fmt.Println("Added to the waiting connections")
			}

		}
	}

}
