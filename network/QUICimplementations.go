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
)

// First we implement the connection methods

func (conn *QUICCon) Connect(ip string, port int) error {
	config := transport.NewConfig()
	config.Params.MaxIdleTimeout = time.Hour
	cert, err := tls.LoadX509KeyPair("network/client.crt", "network/client.key")
	structures.PANIC_ON_ERR(err)
	config.TLS = &tls.Config{}
	config.TLS.InsecureSkipVerify = true
	config.TLS.Certificates = []tls.Certificate{cert}

	client := quic.NewClient(config)
	handler := &quicClientHandler{
		toWrite:            make(chan []byte, 100),
		client:             client,
		connectionAccepted: make(map[string]bool),
		waitForConnection:  sync.NewCond(&sync.Mutex{}),
	}
	conn.handler = handler
	client.SetHandler(handler)
	address := ip + ":" + strconv.Itoa(port)
	err = client.ListenAndServe(ip + ":")
	structures.PANIC_ON_ERR(err)
	err = client.Connect(address)
	go func() {
		structures.PANIC_ON_ERR(client.Serve())
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
	return conn.handler.mainStream.Read(buffer)
}

func (conn *QUICCon) Close() error {
	conn.connection.Close()
	return conn.handler.client.Close()
}

func (conn *QUICCon) RemoteAddr() net.Addr {
	return conn.handler.mainStream.RemoteAddr()
}

func (conn *QUICCon) LocalAddr() net.Addr {
	return conn.handler.mainStream.LocalAddr()
}

// Next the listening methods
func (quicListener *QUICListener) Listen(ip string, port int) error {
	config := transport.NewConfig()
	cert, err := tls.LoadX509KeyPair("network/server.crt", "network/server.key")
	if err != nil {
		return err
	}
	config.Params.MaxIdleTimeout = time.Hour
	config.TLS = &tls.Config{}
	config.TLS.Certificates = []tls.Certificate{cert}
	server := quic.NewServer(config)

	handler := &quicServerHandler{
		waitingConnections:  make(chan *StreamWrapper, 100),
		acceptedConnections: make([]*quic.Conn, 100),
		connectionAccepted:  make(map[string]bool),
		quicServer:          server,
	}
	server.SetHandler(handler)
	quicListener.handler = handler
	go server.ListenAndServe(ip + ":" + strconv.Itoa(port))
	return nil
}

func (quicListener *QUICListener) Close() error {
	return quicListener.handler.quicServer.Close()
}

func (quicListener *QUICListener) Accept() (net.Conn, error) {
	return <-quicListener.handler.waitingConnections, nil
}
