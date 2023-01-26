package network

import (
	"fmt"
	"net"
)

// First we implement the connection methods

func (conn *TCPCon) Connect(ip string, port int) error {
	connection, err := net.Dial("tcp", fmt.Sprint(ip, ":", port))
	if err == nil {
		conn.connection = &connection
	}
	return err
}

func (conn *TCPCon) Write(toWrite []byte) (n int, err error) {
	return (*conn.connection).Write(toWrite)
}

func (conn *TCPCon) Read(buffer []byte) (n int, err error) {
	return (*conn.connection).Read(buffer)
}

func (conn *TCPCon) Close() error {
	return (*conn.connection).Close()
}

func (conn *TCPCon) RemoteAddr() net.Addr {
	return (*conn.connection).RemoteAddr()
}

// Next the listening methods

func (tcpListener *TCPListener) Listen(ip string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprint(ip, ":", port))
	if err == nil {
		tcpListener.listener = &listener
	}
	return err
}

func (tcpListener *TCPListener) Close() error {
	return (*tcpListener.listener).Close()
}

func (tcpListener *TCPListener) Accept() (net.Conn, error) {
	return (*tcpListener.listener).Accept()
}
