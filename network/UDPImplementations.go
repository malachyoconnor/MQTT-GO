package network

import (
	"MQTT-GO/structures"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	udpBufferSize = 1024 * 1024 * 100
)

// Connect implements the Connect function for UDP connections.
// We first dial the address and port, and create a channel for the packet buffer.
// We then start a goroutine that reads from the connection and puts the packets in the buffer.
func (conn *UDPConn) Connect(ip string, port int) error {
	if conn.connected {
		return errors.New("error: Tried to re-open connection")
	}

	serverAddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	// Nil means a 'random' port gets chosen
	connection, err := net.DialUDP("udp", nil, &serverAddr)
	conn.connection = connection

	// 10000 is a magic number
	conn.packetBuffer = make(chan []byte, 10000)

	if err == nil {
		conn.localAddr = connection.LocalAddr()
		conn.remoteAddr = connection.RemoteAddr()
	} else {
		return err
	}

	_ = connection.SetReadBuffer(udpBufferSize)
	_ = connection.SetWriteBuffer(udpBufferSize)
	conn.connectionType = UDPClientConnection

	go clientBackgroundReader(conn)
	conn.connected = true

	return nil
}

func clientBackgroundReader(conn *UDPConn) {
	for {
		buffer := make([]byte, 1024*10)
		bytesRead, receivedAddr, err := conn.connection.ReadFromUDP(buffer)
		buffer = buffer[:bytesRead]

		if errors.Is(err, net.ErrClosed) {
			fmt.Println("Connection closed, disconnecting")
			return
		}

		remoteAddr := *conn.remoteAddr.(*net.UDPAddr)

		if receivedAddr.IP.String() != remoteAddr.IP.String() || receivedAddr.Port != remoteAddr.Port {
			fmt.Println("Received a UDP message from someone else", buffer, receivedAddr.IP, receivedAddr.Port)
			log.Println("Received a UDP message from someone else", buffer, receivedAddr.IP, receivedAddr.Port)
			continue
		}
		if err != nil {
			fmt.Println("ERROR WHILE READING UDP:", err)
			log.Println("ERROR WHILE READING UDP:", err)
		}
		conn.packetBuffer <- buffer
	}
}

func (conn *UDPConn) Write(toWrite []byte) (n int, err error) {
	// If we're a client, we've created a singly connected connection.
	// If we're a server, we've got a general purpose connection with which we can
	// send to multiple addresses.
	conn.writeLock.Lock()
	defer conn.writeLock.Unlock()
	if conn.connectionType == UDPServerConnection {
		return conn.connection.WriteToUDP(toWrite, conn.remoteAddr.(*net.UDPAddr))
	}
	return conn.connection.Write(toWrite)
}

func (conn *UDPConn) Read(buffer []byte) (n int, err error) {
	readData, channelOpen := <-conn.packetBuffer

	if !channelOpen {
		return 0, net.ErrClosed
	}
	return copy(buffer, readData), nil
}

// Close closes the connection.
// If we're a client, we stop listening and close the packet buffer.
func (conn *UDPConn) Close() error {
	// We don't want to close the connection on the other end
	// So we just send a disconnect and stop listening.
	if conn.connectionType == UDPClientConnection {
		for len(conn.packetBuffer) > 0 {
			<-conn.packetBuffer
		}
		close(conn.packetBuffer)
		fmt.Println("Closing a connection from:", conn.localAddr, "to", conn.remoteAddr)
		return conn.connection.Close()
	}

	if conn.connectionType == UDPServerConnection {
		conn.serverConnectionDeleter()
		close(conn.packetBuffer)
		conn.connected = false
	}
	return nil
}

// RemoteAddr returns the remote address of the connection.
func (conn *UDPConn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

// LocalAddr returns the local address of the connection.
func (conn *UDPConn) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

// SetDeadline sets the deadline associated with the connection.
func (conn *UDPConn) SetDeadline(t time.Time) error {
	return conn.connection.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
func (conn *UDPConn) SetReadDeadline(t time.Time) error {
	return conn.connection.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
func (conn *UDPConn) SetWriteDeadline(t time.Time) error {
	return conn.connection.SetWriteDeadline(t)
}

// Next the listening methods

// Listen starts listening on the given port. We create and open connections map, and a new client buffer.
// We then start a goroutine that listens for new connections, updates the open connections map
// and puts them in the buffer.
func (udpListener *UDPListener) Listen(ip string, port int) error {
	if udpListener.listening {
		return errors.New("error: Attempted to listen when already listening")
	}

	udpListener.openConnections = structures.CreateSafeMap[string, chan []byte]()
	// We can buffer 300 new clients before having to clear them
	udpListener.newClientBuffer = make(chan net.Addr, 300)
	laddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	udpListener.localAddr = &laddr
	connection, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		return err
	}

	err = connection.SetReadBuffer(udpBufferSize)
	if err != nil {
		return err
	}
	err = connection.SetWriteBuffer(udpBufferSize)
	if err != nil {
		return err
	}
	go startUDPbackgroundListener(udpListener, connection)
	udpListener.listener = connection

	udpListener.listening = true
	return err
}

type udpMessage struct {
	buffer []byte
	addr   *net.UDPAddr
}

func startUDPbackgroundListener(udpListener *UDPListener, connection *net.UDPConn) {
	readMessageBufferEven := make(chan *udpMessage, 1000)
	readMessageBufferOdd := make(chan *udpMessage, 1000)

	go handleMessageForwarding(readMessageBufferEven, udpListener)
	go handleMessageForwarding(readMessageBufferOdd, udpListener)

	_ = connection.SetWriteBuffer(1024 * 1024 * 50)
	connection.SetReadBuffer(1024 * 1024 * 50)

	for {
		buffer := make([]byte, 1024*20)
		bytesRead, receivedAddr, err := connection.ReadFromUDP(buffer)
		buffer = buffer[:bytesRead]
		if errors.Is(err, net.ErrClosed) {
			fmt.Println("Connection closed, ceasing to listen")
			return
		}
		if err != nil {
			fmt.Println(err)
		}

		if receivedAddr.IP[0]%2 == 0 {
			readMessageBufferEven <- &udpMessage{buffer: buffer, addr: receivedAddr}
		} else {
			readMessageBufferOdd <- &udpMessage{buffer: buffer, addr: receivedAddr}
		}
	}
}

func handleMessageForwarding(readMessageBuffer chan *udpMessage, udpListener *UDPListener) {
	for {
		udpMsg := <-readMessageBuffer
		packet := udpMsg.buffer
		receivedAddr := udpMsg.addr

		address := fmt.Sprint(receivedAddr.IP, ":", receivedAddr.Port)
		udpListener.openConnectionsLock.Lock()
		if !udpListener.openConnections.Contains(address) {
			udpListener.openConnections.Put(address, make(chan []byte, 2000))
			udpListener.newClientBuffer <- receivedAddr
		}
		udpListener.openConnections.Get(address) <- packet
		udpListener.openConnectionsLock.Unlock()
	}
}

// Close closes the listener.
func (udpListener *UDPListener) Close() error {
	return udpListener.listener.Close()
}

// Note that UDP is connectionless, so we need to create our own connections. That means listening for ALL UDP packets
// and pushing them down the correct connection.

// Accept waits connections from the newClientBuffer from the background listener, and returns a new connection.
func (udpListener *UDPListener) Accept() (Conn, error) {
	newClientAddress := <-udpListener.newClientBuffer
	newClientUDPAddress := newClientAddress.(*net.UDPAddr)
	stringAddress := fmt.Sprint(newClientUDPAddress.IP, ":", newClientUDPAddress.Port)

	udpListener.openConnectionsLock.RLock()

	newConnection := UDPConn{
		packetBuffer:   udpListener.openConnections.Get(stringAddress),
		remoteAddr:     newClientAddress,
		localAddr:      udpListener.listener.LocalAddr(),
		connection:     udpListener.listener,
		connected:      true,
		connectionType: UDPServerConnection,
		writeLock:      &sync.Mutex{},
		// We pass a function which can DELETE an open connection upon a disconnect.
		serverConnectionDeleter: func() {
			udpListener.openConnections.Delete(stringAddress)
		},
	}
	udpListener.openConnectionsLock.RUnlock()

	return &newConnection, nil
}
