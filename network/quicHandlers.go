package network

import (
	"MQTT-GO/structures"
	"net"
	"sync"
	"time"

	"github.com/goburrow/quic"
	"github.com/goburrow/quic/transport"
)

type quicClientHandler struct {
	toWrite    chan []byte
	mainStream *StreamWrapper
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

				handler.mainStream = &StreamWrapper{
					readLock:  sync.Mutex{},
					writeLock: sync.Mutex{},
					stream:    st,
				}

			}
		}
	}
}

type quicServerHandler struct {
	waitingConnections  chan *StreamWrapper
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
					structures.Println("Error while connecting to stream:", err)
					continue
				}

				handler.waitingConnections <- &StreamWrapper{
					readLock:  sync.Mutex{},
					writeLock: sync.Mutex{},
					stream:    st,
				}
				structures.Println("Added to the waiting connections")
			}

		}
	}

}

type StreamWrapper struct {
	stream    *quic.Stream
	readLock  sync.Mutex
	writeLock sync.Mutex
}

func (wrapper *StreamWrapper) Write(b []byte) (int, error) {
	wrapper.writeLock.Lock()
	structures.Println("Started writing!", b)
	defer structures.Println("Finished writing!")

	defer wrapper.writeLock.Unlock()
	return wrapper.stream.Write(b)
}

func (wrapper *StreamWrapper) Read(b []byte) (int, error) {
	wrapper.readLock.Lock()
	structures.Println("Started Reading!")
	defer structures.Println("Finished Reading!")
	defer wrapper.readLock.Unlock()
	return wrapper.stream.Read(b)
}

func (wrapper *StreamWrapper) Close() error {
	wrapper.writeLock.Lock()
	structures.Println("Started Closing!")
	defer structures.Println("Finished Closing!")
	defer wrapper.writeLock.Unlock()
	wrapper.stream.CloseRead(0)
	wrapper.stream.CloseWrite(0)
	return wrapper.stream.Close()
}

func (wrapper *StreamWrapper) LocalAddr() net.Addr {
	wrapper.readLock.Lock()
	structures.Println("Started LocalAddressing!")
	defer structures.Println("Finished LocalAddressing!")
	defer wrapper.readLock.Unlock()
	return wrapper.stream.LocalAddr()
}

func (wrapper *StreamWrapper) RemoteAddr() net.Addr {
	wrapper.readLock.Lock()
	defer wrapper.readLock.Unlock()
	return wrapper.stream.RemoteAddr()
}

func (wrapper *StreamWrapper) SetDeadline(t time.Time) error {
	wrapper.readLock.Lock()
	defer wrapper.readLock.Unlock()
	return wrapper.stream.SetDeadline(t)
}

func (wrapper *StreamWrapper) SetReadDeadline(t time.Time) error {
	wrapper.readLock.Lock()
	defer wrapper.readLock.Unlock()
	return wrapper.stream.SetReadDeadline(t)
}

func (wrapper *StreamWrapper) SetWriteDeadline(t time.Time) error {
	wrapper.readLock.Lock()
	defer wrapper.readLock.Unlock()
	return wrapper.stream.SetWriteDeadline(t)
}
