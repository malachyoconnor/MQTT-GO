package network

import (
	"MQTT-GO/structures"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goburrow/quic"
	"github.com/goburrow/quic/transport"
	"github.com/sasha-s/go-deadlock"
)

type quicClientHandler struct {
	toWrite            chan []byte
	mainStream         *StreamWrapper
	client             *quic.Client
	connectionAccepted map[string]bool
	waitForConnection  *deadlock.Cond
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
				streamID, ok := conn.NewStream(true)
				if !ok {
					structures.Println("Error can't open stream")
					return
				}
				st, err := conn.Stream(streamID)
				if err != nil {
					panic(err)
				}
				// Once we've opened the stream, tell the waiting readers/writers that
				// they can access it
				handler.waitForConnection.L.Lock()
				handler.mainStream = &StreamWrapper{
					readLock:   deadlock.Mutex{},
					writeLock:  deadlock.Mutex{},
					stream:     st,
					output:     false,
					connection: conn,
				}
				handler.waitForConnection.Broadcast()
				handler.waitForConnection.L.Unlock()

			}
		}
	}
}

type quicServerHandler struct {
	waitingConnections  chan *StreamWrapper
	acceptedConnections []*StreamWrapper
	connectionAccepted  *structures.SafeMap[string, *StreamWrapper]
	quicServer          *quic.Server
}

var (
	closed atomic.Int32
)

func (handler *quicServerHandler) Serve(conn *quic.Conn, events []transport.Event) {
	if handler.connectionAccepted.Contains(conn.RemoteAddr().String()) {
		handler.connectionAccepted.Get(conn.RemoteAddr().String()).connectionLock.Lock()
		defer handler.connectionAccepted.Get(conn.RemoteAddr().String()).connectionLock.Unlock()
	}
	for _, e := range events {
		switch e.Type {

		case transport.EventStreamOpen:
			{
				if handler.connectionAccepted.Contains(conn.RemoteAddr().String()) {
					continue
				}
				st, err := conn.Stream(e.Data)
				if err != nil {
					structures.Println("Error while connecting to stream:", err)
					continue
				}

				streamWrapper := &StreamWrapper{
					connection: conn,
					readLock:   deadlock.Mutex{},
					writeLock:  deadlock.Mutex{},
					stream:     st,
					output:     false,
				}
				handler.connectionAccepted.Put(conn.RemoteAddr().String(), streamWrapper)
				handler.waitingConnections <- streamWrapper
			}
		case transport.EventConnClosed:
			{
				structures.PrintCentrally("CONNECTION CLOSED")
				handler.connectionAccepted.Delete(conn.RemoteAddr().String())
				conn.Close()
				structures.Println("Closed", closed.Add(1))
			}
		}

	}

}

type StreamWrapper struct {
	stream         *quic.Stream
	readLock       deadlock.Mutex
	writeLock      deadlock.Mutex
	output         bool
	connection     *quic.Conn
	connectionLock sync.Mutex
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
	defer wrapper.writeLock.Unlock()
	if wrapper.output {
		structures.Println("Started Closing!")
		defer structures.Println("Finished Closing!")
	}
	wrapper.connectionLock.Lock()
	defer wrapper.connectionLock.Unlock()
	if wrapper.connection.ConnectionState().State == "closed" {
		return nil
	}
	return wrapper.connection.Close()
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
