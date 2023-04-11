package network

import (
	"crypto/tls"
	"net"
	"strconv"

	"github.com/goburrow/quic"
	"github.com/goburrow/quic/transport"
)

// First we implement the connection methods

func (conn *QUICCon) Connect(ip string, port int) error {
	config := transport.NewConfig()
	cert, err := tls.LoadX509KeyPair("C:\\Users\\malac\\Desktop\\MQTT-GO\\network\\client.crt",
		"C:\\Users\\malac\\Desktop\\MQTT-GO\\network\\client.key")
	if err != nil {
		panic(err)
	}
	config.TLS = &tls.Config{}
	config.TLS.InsecureSkipVerify = true
	config.TLS.Certificates = []tls.Certificate{cert}

	client := quic.NewClient(config)
	handler := &quicClientHandler{
		toWrite: make(chan []byte, 100),
		client:  client,
	}
	conn.handler = handler
	client.SetHandler(handler)
	address := ip + ":" + strconv.Itoa(port)
	err = client.ListenAndServe(ip + ":")

	if err != nil {
		panic(err)
	}
	err = client.Connect(address)
	go client.Serve()

	if err != nil {
		panic(err)
	}

	for conn.handler.mainStream == nil {
	}
	conn.connection = conn.handler.mainStream

	return nil
}

func (conn *QUICCon) Write(toWrite []byte) (n int, err error) {
	for conn.handler.mainStream == nil {
	}
	return conn.handler.mainStream.Write(toWrite)
}

func (conn *QUICCon) Read(buffer []byte) (n int, err error) {
	return conn.handler.mainStream.Read(buffer)
}

func (conn *QUICCon) Close() error {
	return conn.handler.client.Close()
}

func (conn *QUICCon) RemoteAddr() net.Addr {
	return conn.handler.mainStream.RemoteAddr()
}

// Next the listening methods
func (quicListener *QUICListener) Listen(ip string, port int) error {
	config := transport.NewConfig()
	cert, err := tls.LoadX509KeyPair("C:\\Users\\malac\\Desktop\\MQTT-GO\\network\\server.crt",
		"C:\\Users\\malac\\Desktop\\MQTT-GO\\network\\server.key")
	if err != nil {
		return err
	}
	config.TLS = &tls.Config{}
	config.TLS.Certificates = []tls.Certificate{cert}
	server := quic.NewServer(config)

	handler := &quicServerHandler{
		waitingConnections:  make(chan *quic.Stream, 20),
		acceptedConnections: make([]*quic.Conn, 100),
		connectionAccepted:  make(map[net.Addr]bool),
		quicServer:          server,
	}
	server.SetHandler(handler)
	quicListener.handler = handler
	go server.ListenAndServe(ip + ":" + strconv.Itoa(port))
	return nil
}

func (quicListener *QUICListener) Close() error {
	return nil
}

func (quicListener *QUICListener) Accept() (net.Conn, error) {
	return <-quicListener.handler.waitingConnections, nil
}
