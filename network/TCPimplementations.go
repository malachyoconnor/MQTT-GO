package network

import (
	"fmt"
	"net"
	"time"
)

// First we implement the connection methods

// Connect implements the Connect function for TCP connections.
func (conn *TCPConn) Connect(ip string, port int) error {

	connection, err := net.Dial("tcp", fmt.Sprint(ip, ":", port))
	if err == nil {
		conn.connection = connection.(*net.TCPConn)
	}
	return err
}

// Write writes to the TCP connection associated with the TCPConn.
func (conn *TCPConn) Write(toWrite []byte) (n int, err error) {
	return (*conn.connection).Write(toWrite)
}

// Read reads from the TCP connection associated with the TCPConn.
func (conn *TCPConn) Read(buffer []byte) (n int, err error) {
	return (*conn.connection).Read(buffer)
}

// Close closes the TCP connection associated with the TCPConn.
func (conn *TCPConn) Close() error {
	return (*conn.connection).Close()
}

// RemoteAddr returns the remote address of the TCP connection associated with the TCPConn.
func (conn *TCPConn) RemoteAddr() net.Addr {
	return (*conn.connection).RemoteAddr()
}

// LocalAddr returns the local address of the TCP connection associated with the TCPConn.
func (conn *TCPConn) LocalAddr() net.Addr {
	return (*conn.connection).LocalAddr()
}

func (conn *TCPConn) SetDeadline(t time.Time) error {
	return (*conn.connection).SetDeadline(t)
}

func (conn *TCPConn) SetReadDeadline(t time.Time) error {
	return (*conn.connection).SetReadDeadline(t)
}

func (conn *TCPConn) SetWriteDeadline(t time.Time) error {
	return (*conn.connection).SetWriteDeadline(t)
}

// Next the listening methods

// Listen implements the Listen function for TCP connections.
func (tcpListener *TCPListener) Listen(ip string, port int) error {
	tcpAddr := &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	if err == nil {
		tcpListener.listener = listener
	}
	return err
}

// Close closes the TCP listener.
func (tcpListener *TCPListener) Close() error {
	return (*tcpListener.listener).Close()
}

// Accept accepts a TCP connection from the TCP listener.
func (tcpListener *TCPListener) Accept() (Conn, error) {
	connection, err := (*tcpListener.listener).AcceptTCP()
	if err != nil {
		return nil, err
	}
	return &TCPConn{
		connection: connection,
	}, err
}
