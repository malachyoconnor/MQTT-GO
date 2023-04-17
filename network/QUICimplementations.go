package network

import (
	"MQTT-GO/structures"
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/goburrow/quic"
	"github.com/goburrow/quic/transport"
	"github.com/sasha-s/go-deadlock"
)

// First we implement the connection methods

func (conn *QUICCon) Connect(ip string, port int) error {
	config := transport.NewConfig()
	config.Params.MaxIdleTimeout = time.Hour
	config.Params.InitialMaxStreamDataBidiLocal = 2
	config.Params.InitialMaxStreamDataUni = 2
	config.Params.InitialMaxStreamsUni = 2

	cert, err := tls.LoadX509KeyPair("network/client.crt", "network/client.key")
	structures.PANIC_ON_ERR(err)
	config.TLS = &tls.Config{}
	config.TLS.InsecureSkipVerify = true
	config.TLS.Certificates = []tls.Certificate{cert}
	client := quic.NewClient(config)

	handler := &quicClientHandler{
		toWrite:            make(chan []byte, 2000),
		client:             client,
		connectionAccepted: make(map[string]bool),
		waitForConnection:  &deadlock.Cond{Cond: *sync.NewCond(&sync.Mutex{})},
	}
	conn.handler = handler
	client.SetHandler(handler)
	address := ip + ":" + strconv.Itoa(port)
	err = client.ListenAndServe(ip + ":")
	structures.PANIC_ON_ERR(err)
	err = client.Connect(address)
	go func() {
		defer client.Close()
		client.Serve()
	}()
	structures.PANIC_ON_ERR(err)

	conn.handler.waitForConnection.L.Lock()
	if conn.handler.mainStream == nil {
		conn.handler.waitForConnection.Wait()
	}
	conn.handler.waitForConnection.L.Unlock()
	conn.connection = conn.handler.mainStream
	return nil
}

func (conn *QUICCon) Write(toWrite []byte) (n int, err error) {
	for conn.handler.mainStream == nil {
		// structures.Print("\rSpinning")
		time.Sleep(time.Millisecond)
	}
	return conn.handler.mainStream.Write(toWrite)
}

func (conn *QUICCon) Read(buffer []byte) (n int, err error) {
	for conn.handler.mainStream == nil {
		time.Sleep(100 * time.Millisecond)
	}
	return conn.handler.mainStream.Read(buffer)
}

func (conn *QUICCon) Close() error {
	return conn.handler.client.Close()
}

func (conn *QUICCon) RemoteAddr() net.Addr {
	return conn.handler.mainStream.RemoteAddr()
}

func (conn *QUICCon) LocalAddr() net.Addr {
	return conn.handler.mainStream.LocalAddr()
}

var (
// handlerCreated = false
// handler        = &quicServerHandler{}
)

// Next the listening methods
func (quicListener *QUICListener) Listen(ip string, port int) error {
	config := transport.NewConfig()
	cert, err := tls.LoadX509KeyPair("network/server.crt", "network/server.key")
	if err != nil {
		return err
	}
	config.Params.InitialMaxStreamsUni = 2
	config.Params.InitialMaxStreamDataBidiLocal = 2
	config.Params.InitialMaxStreamDataUni = 2
	config.Params.MaxIdleTimeout = time.Hour
	config.Params.MaxAckDelay = time.Second
	config.TLS = &tls.Config{}
	config.TLS.Certificates = []tls.Certificate{cert}
	server := quic.NewServer(config)

	structures.PrintCentrally("MAKING NEW HANDLER")
	handler := &quicServerHandler{
		waitingConnections:  make(chan *StreamWrapper, 2000),
		acceptedConnections: make([]*StreamWrapper, 0),
		connectionAccepted:  structures.CreateSafeMap[string, *StreamWrapper](),
		quicServer:          server,
	}
	server.SetHandler(handler)
	quicListener.handler = handler
	go func() {
		defer server.Close()
		server.ListenAndServe(ip + ":" + strconv.Itoa(port))
	}()
	return nil
}

func (quicListener *QUICListener) Close() error {
	for _, conn := range quicListener.handler.acceptedConnections {
		structures.PANIC_ON_ERR(conn.Close())
	}

	return quicListener.handler.quicServer.Close()
}

// This works!
func (quicListener *QUICListener) Accept() (net.Conn, error) {
	acceptedConnection := <-quicListener.handler.waitingConnections
	quicListener.handler.acceptedConnections = append(quicListener.handler.acceptedConnections, acceptedConnection)
	return acceptedConnection, nil
}
