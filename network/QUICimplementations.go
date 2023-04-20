package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
)

// First we implement the connection methods for QUIC

// Connect implements the Connect function for QUIC connections.
// We first load a certificate and key, and create a tls.Config with the certificate.
// We then create a quic.Config with a MaxIdleTimeout of 1 hour.
// Finally, we dial the address and port, and open a stream.
func (conn *QUICCon) Connect(ip string, port int) error {
	cert, err := tls.LoadX509KeyPair("network/client.crt", "network/client.key")
	if err != nil {
		return err
	}
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	config.NextProtos = []string{"UDP"}
	config.InsecureSkipVerify = true
	config.Certificates = []tls.Certificate{cert}

	quicConfig := &quic.Config{}
	quicConfig.MaxIdleTimeout = time.Hour

	connection, err := quic.DialAddr(ip+":"+fmt.Sprint(port), config, quicConfig)

	if err != nil {
		fmt.Println("ERROR WHILE DIALING")
		return err
	}

	conn.connection = &connection
	stream, err := connection.OpenStream()
	if err != nil {
		return err
	}
	conn.stream = &stream
	return nil
}

// Write writes to the QUIC Stream associated with the QUICCon.
func (conn *QUICCon) Write(toWrite []byte) (n int, err error) {
	conn.streamWriteLock.Lock()
	defer conn.streamWriteLock.Unlock()
	return (*conn.stream).Write(toWrite)
}

// Read reads from the QUIC Stream associated with the QUICCon.
func (conn *QUICCon) Read(buffer []byte) (n int, err error) {
	conn.streamReadLock.Lock()
	defer conn.streamReadLock.Unlock()
	return (*conn.stream).Read(buffer)
}

// Close closes the QUIC Stream and the QUIC Connection associated with the QUICCon.
// It closes with the "error" bye.
func (conn *QUICCon) Close() error {
	err := quic.ApplicationErrorCode(0)
	return (*conn.connection).CloseWithError(err, "bye")
}

// RemoteAddr returns the remote address of the QUIC Connection associated with the QUICCon.
func (conn *QUICCon) RemoteAddr() net.Addr {
	return (*conn.connection).RemoteAddr()
}

// LocalAddr returns the local address of the QUIC Connection associated with the QUICCon.
func (conn *QUICCon) LocalAddr() net.Addr {
	return (*conn.connection).LocalAddr()
}

// SetDeadline sets the deadline for the QUIC Stream associated with the QUICCon.
func (conn *QUICCon) SetDeadline(t time.Time) error {
	return (*conn.stream).SetDeadline(t)
}

// SetReadDeadline sets the read deadline for the QUIC Stream associated with the QUICCon.
func (conn *QUICCon) SetReadDeadline(t time.Time) error {
	return (*conn.stream).SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline for the QUIC Stream associated with the QUICCon.
func (conn *QUICCon) SetWriteDeadline(t time.Time) error {
	return (*conn.stream).SetWriteDeadline(t)
}

// Listen is a function that implements the Listen function for QUIC connections.
// We first load a certificate and key, and create a tls.Config with the certificate.
// We then create a quic.Config with a MaxIdleTimeout of 1 hour.
// Finally, we listen on the address and port, and open a stream.
func (quicListener *QUICListener) Listen(ip string, port int) error {
	laddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	connection, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		return err
	}
	cert, err := tls.LoadX509KeyPair("network/server.crt", "network/server.key")

	if err != nil {
		return err
	}

	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	config.NextProtos = []string{"UDP"}
	config.Certificates = []tls.Certificate{cert}
	config.InsecureSkipVerify = true
	quicConfig := &quic.Config{}
	quicConfig.MaxIdleTimeout = time.Hour

	listener, err := quic.Listen(connection, config, quicConfig)
	quicListener.listener = &listener

	if err != nil {
		return err
	}

	return nil
}

// Close closes the QUIC Listener.
func (quicListener *QUICListener) Close() error {
	return (*quicListener.listener).Close()
}

// Accept accepts a QUIC connection and returns a QUICCon.
// It first accepts a QUIC connection, and then a QUIC stream.
func (quicListener *QUICListener) Accept() (net.Conn, error) {
	context := context.Background()
	conn, err := (*quicListener.listener).Accept(context)
	if err != nil {
		return nil, err
	}

	stream, err := conn.AcceptStream(context)
	if err != nil {
		return nil, err
	}

	return &QUICCon{
		connection:      &conn,
		stream:          &stream,
		streamReadLock:  &sync.Mutex{},
		streamWriteLock: &sync.Mutex{},
	}, nil
}
