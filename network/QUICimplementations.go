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

// First we implement the connection methods

func (conn *QUICCon) Connect(ip string, port int) error {

	cert, err := tls.LoadX509KeyPair("network/client.crt", "network/client.key")
	if err != nil {
		return err
	}
	config := &tls.Config{}
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

func (conn *QUICCon) Write(toWrite []byte) (n int, err error) {
	conn.streamWriteLock.Lock()
	defer conn.streamWriteLock.Unlock()
	return (*conn.stream).Write(toWrite)
}

func (conn *QUICCon) Read(buffer []byte) (n int, err error) {
	conn.streamReadLock.Lock()
	defer conn.streamReadLock.Unlock()
	return (*conn.stream).Read(buffer)
}

func (conn *QUICCon) Close() error {
	err := quic.ApplicationErrorCode(0)
	return (*conn.connection).CloseWithError(err, "bye")
}

func (conn *QUICCon) RemoteAddr() net.Addr {
	return (*conn.connection).RemoteAddr()
}

func (conn *QUICCon) LocalAddr() net.Addr {
	return (*conn.connection).LocalAddr()
}

func (conn *QUICCon) SetDeadline(t time.Time) error {
	return (*conn.stream).SetDeadline(t)
}

func (conn *QUICCon) SetReadDeadline(t time.Time) error {
	return (*conn.stream).SetReadDeadline(t)
}

func (conn *QUICCon) SetWriteDeadline(t time.Time) error {
	return (*conn.stream).SetWriteDeadline(t)
}

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

	config := &tls.Config{}
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

func (quicListener *QUICListener) Close() error {
	return (*quicListener.listener).Close()
}

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
