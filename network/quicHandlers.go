package network

import (
	"MQTT-GO/structures"
	"net"
	"sync"
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
	connectionAccepted  *structures.SafeMap[string, bool]
	quicServer          *quic.Server
}

func (handler *quicServerHandler) Serve(conn *quic.Conn, events []transport.Event) {

	for _, e := range events {
		switch e.Type {

		case transport.EventStreamOpen:
			{

				if handler.connectionAccepted.Get(conn.RemoteAddr().String()) {
					continue
				}
				st, err := conn.Stream(e.Data)
				if err != nil {
					structures.Println("Error while connecting to stream:", err)
					continue
				}
				handler.connectionAccepted.Put(conn.RemoteAddr().String(), true)
				handler.waitingConnections <- &StreamWrapper{
					connection: conn,
					readLock:   deadlock.Mutex{},
					writeLock:  deadlock.Mutex{},
					stream:     st,
					output:     false,
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
	stream     *quic.Stream
	readLock   deadlock.Mutex
	writeLock  deadlock.Mutex
	output     bool
	connection *quic.Conn
}

func (wrapper *StreamWrapper) Write(b []byte) (int, error) {
	wrapper.writeLock.Lock()
	if wrapper.output {
		structures.Println("Started writing!", b)
		defer structures.Println("Finished writing!")
	}
	defer wrapper.writeLock.Unlock()

	// wrapper.stream.SetWriteDeadline(time.Now().Add(time.Second * 5))
	n, err := wrapper.stream.Write(b)

	// for err != nil {
	// 	structures.PrintCentrally(err)
	// 	n, err = wrapper.stream.Write(b)
	// }

	return n, err
}

// TRY REPEATING THIS UNTIL IT DOESNT GIVE IO TIMEOUT

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

// serverHandler := func(conn *Conn, events []transport.Event) {
// 	// t.Logf("server events: cid=%x %v", conn.scid, events)
// 	for _, e := range events {
// 		switch e.Type {
// 		case transport.EventStreamOpen:
// 			st, err := conn.Stream(e.Data)
// 			if err != nil {
// 				t.Errorf("server stream %v: %v", e.Data, err)
// 				conn.Close()
// 				return
// 			}
// 			go recvFn(st)
// 		}
// 	}
// }
// clientHandler := func(conn *Conn, events []transport.Event) {
// 	// t.Logf("client events: cid=%x %v", conn.scid, events)
// 	for _, e := range events {
// 		switch e.Type {
// 		case transport.EventConnOpen:
// 			id, ok := conn.NewStream(true)
// 			if !ok {
// 				t.Error("client newstream failed")
// 				conn.Close()
// 				return
// 			}
// 			st, err := conn.Stream(id)
// 			if err != nil {
// 				t.Errorf("client stream: %d %v", id, err)
// 				conn.Close()
// 				return
// 			}
// 			go sendFn(st)
// 		}
// 	}
// }
