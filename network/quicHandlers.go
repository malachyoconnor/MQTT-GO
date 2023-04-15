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
	toWrite            chan []byte
	mainStream         *StreamWrapper
	client             *quic.Client
	connectionAccepted map[string]bool
	waitForConnection  *sync.Cond
}

func (handler *quicClientHandler) Serve(conn *quic.Conn, events []transport.Event) {

	for _, e := range events {

		switch e.Type {
		case transport.EventConnOpen:
			{
				if handler.connectionAccepted[conn.RemoteAddr().String()] {
					break
				}
				handler.connectionAccepted[conn.RemoteAddr().String()] = true
				st, err := conn.Stream(e.Data)
				if err != nil {
					panic(err)
				}
				// Once we've opened the stream, tell the waiting readers/writers that
				// they can access it
				handler.waitForConnection.L.Lock()
				handler.mainStream = &StreamWrapper{
					readLock:  sync.Mutex{},
					writeLock: sync.Mutex{},
					stream:    st,
					output:    false,
				}
				handler.waitForConnection.Broadcast()
				handler.waitForConnection.L.Unlock()

			}
		}
	}
}

type quicServerHandler struct {
	waitingConnections  chan *StreamWrapper
	acceptedConnections []*quic.Conn
	connectionAccepted  map[string]bool
	quicServer          *quic.Server
}

func (handler *quicServerHandler) Serve(conn *quic.Conn, events []transport.Event) {

	for _, e := range events {
		switch e.Type {

		case transport.EventStreamOpen:
			{
				if handler.connectionAccepted[conn.LocalAddr().String()] {
					break
				}
				st, err := conn.Stream(e.Data)
				if err != nil {
					structures.Println("Error while connecting to stream:", err)
					continue
				}
				handler.connectionAccepted[conn.LocalAddr().String()] = true
				handler.waitingConnections <- &StreamWrapper{
					readLock:  sync.Mutex{},
					writeLock: sync.Mutex{},
					stream:    st,
					output:    false,
				}
			}
		case transport.EventConnClosed:
			{
				conn.Close()
				structures.PrintCentrally("CONNECTION CLOSED")
			}
		}

	}

}

type StreamWrapper struct {
	stream    *quic.Stream
	readLock  sync.Mutex
	writeLock sync.Mutex
	output    bool
}

func (wrapper *StreamWrapper) Write(b []byte) (int, error) {
	wrapper.writeLock.Lock()
	if wrapper.output {
		structures.Println("Started writing!", b)
		defer structures.Println("Finished writing!")
	}
	defer wrapper.writeLock.Unlock()
	return wrapper.stream.Write(b)
}

func (wrapper *StreamWrapper) Read(b []byte) (int, error) {
	wrapper.readLock.Lock()
	if wrapper.output {
		structures.Println("Started Reading!")
		defer structures.Println("Finished Reading!")
	}
	defer wrapper.readLock.Unlock()
	return wrapper.stream.Read(b)
}

func (wrapper *StreamWrapper) Close() error {
	wrapper.writeLock.Lock()
	if wrapper.output {
		structures.Println("Started Closing!")
		defer structures.Println("Finished Closing!")
	}
	defer wrapper.writeLock.Unlock()
	wrapper.stream.CloseRead(0)
	wrapper.stream.CloseWrite(0)
	return wrapper.stream.Close()
}

func (wrapper *StreamWrapper) LocalAddr() net.Addr {
	return wrapper.stream.LocalAddr()
}

func (wrapper *StreamWrapper) RemoteAddr() net.Addr {
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
