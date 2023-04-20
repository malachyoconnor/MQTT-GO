package network

import (
	"MQTT-GO/structures"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// Connect implements the Connect function for UDP connections.
// We first dial the address and port, and create a channel for the packet buffer.
// We then start a goroutine that reads from the connection and puts the packets in the buffer.
func (conn *UDPCon) Connect(ip string, port int) error {
	if conn.connected {
		return errors.New("error: Tried to re-open connection")
	}
	// Parsing localhost would otherwise return nil
	if ip == "localhost" {
		ip = "127.0.0.1"
	}

	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	// Nil means a 'random' port gets chosen
	connection, err := net.DialUDP("udp", nil, &remoteAddr)
	conn.connection = connection
	// 100 is a magic number
	conn.packetBuffer = make(chan []byte, 100)

	if err == nil {
		conn.localAddr = connection.LocalAddr().String()
		conn.remoteAddr = connection.RemoteAddr().String()
	} else {
		return err
	}
	conn.connectionType = UDPClientConnection

	go func() {
		for {
			buffer := make([]byte, 1024)
			bytesRead, receivedAddr, err := conn.connection.ReadFromUDP(buffer)
			buffer = buffer[:bytesRead]

			if errors.Is(err, net.ErrClosed) {
				structures.Println("Connection closed, disconnecting")
				return
			}

			if receivedAddr.IP.String() != remoteAddr.IP.String() || receivedAddr.Port != remoteAddr.Port {
				structures.Println("Received a UDP message from someone else", buffer, receivedAddr.IP, receivedAddr.Port)
				log.Println("Received a UDP message from someone else", buffer, receivedAddr.IP, receivedAddr.Port)
				continue
			}
			if err != nil {
				structures.Println("ERROR WHILE READING UDP:", err)
				log.Println("ERROR WHILE READING UDP:", err)
			}
			conn.packetBuffer <- buffer
		}
	}()
	conn.connected = true

	return nil
}

func (conn *UDPCon) Write(toWrite []byte) (n int, err error) {
	// If we're a client, we've created a singly connected connection.
	// If we're a server, we've got a general purpose connection with which we can
	// send to multiple addresses.
	if conn.connectionType == UDPServerConnection {
		return conn.connection.WriteToUDP(toWrite, conn.RemoteAddr().(*net.UDPAddr))
	}
	return conn.connection.Write(toWrite)
}

func (conn *UDPCon) Read(buffer []byte) (n int, err error) {
	readData, channelOpen := <-conn.packetBuffer

	if !channelOpen {
		return 0, net.ErrClosed
	}
	return copy(buffer, readData), nil
}

// Close closes the connection.
// If we're a client, we stop listening and close the packet buffer.
func (conn *UDPCon) Close() error {
	// We don't want to close the connection on the other end
	// So we just send a disconnect and stop listening.
	if conn.connectionType == UDPClientConnection {
		close(conn.packetBuffer)
		structures.Println("Closing a connection from:", conn.localAddr, "to", conn.remoteAddr)
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
func (conn *UDPCon) RemoteAddr() net.Addr {
	remoteAddr, err := net.ResolveUDPAddr("udp", conn.remoteAddr)
	if err != nil {
		structures.Println("Error while resolving UDP address:", err)
		return nil
	}
	return remoteAddr
}

// LocalAddr returns the local address of the connection.
func (conn *UDPCon) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

// SetDeadline sets the deadline associated with the connection.
func (conn *UDPCon) SetDeadline(t time.Time) error {
	return conn.connection.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
func (conn *UDPCon) SetReadDeadline(t time.Time) error {
	return conn.connection.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
func (conn *UDPCon) SetWriteDeadline(t time.Time) error {
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

	udpListener.openConnections = make(map[string]chan []byte)
	// We can buffer 100 new clients before having to clear them
	udpListener.newClientBuffer = make(chan string, 100)
	laddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	udpListener.localAddr = &laddr
	connection, err := net.ListenUDP("udp", &laddr)
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
	readMessageBufferEven := make(chan *udpMessage, 100)
	readMessageBufferOdd := make(chan *udpMessage, 100)

	go handleReadMessage(readMessageBufferEven, udpListener)
	go handleReadMessage(readMessageBufferOdd, udpListener)

	for {
		buffer := make([]byte, 1024)
		bytesRead, receivedAddr, err := connection.ReadFromUDP(buffer)
		if errors.Is(err, net.ErrClosed) {
			structures.Println("Connection closed, ceasing to listen")
			return
		}
		if err != nil {
			structures.Println(err)
		}

		if receivedAddr.IP[0]%2 == 0 {
			readMessageBufferEven <- &udpMessage{buffer: buffer[:bytesRead], addr: receivedAddr}
		} else {
			readMessageBufferOdd <- &udpMessage{buffer: buffer[:bytesRead], addr: receivedAddr}
		}
	}
}

func handleReadMessage(readMessageBuffer chan *udpMessage, udpListener *UDPListener) {
	for {
		udpMsg := <-readMessageBuffer
		buffer := udpMsg.buffer
		receivedAddr := udpMsg.addr

		address := fmt.Sprint(receivedAddr.IP.String(), ":", receivedAddr.Port)
		udpListener.openConnectionsLock.Lock()
		if _, found := udpListener.openConnections[address]; !found {
			udpListener.openConnections[address] = make(chan []byte, 512)
			udpListener.openConnections[address] <- buffer
			udpListener.newClientBuffer <- address
		}
		udpListener.openConnections[address] <- buffer
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
func (udpListener *UDPListener) Accept() (net.Conn, error) {
	newClientAddress := <-udpListener.newClientBuffer

	udpListener.openConnectionsLock.RLock()

	newConnection := UDPCon{
		packetBuffer:   udpListener.openConnections[newClientAddress],
		remoteAddr:     newClientAddress,
		localAddr:      udpListener.listener.LocalAddr().String(),
		connection:     udpListener.listener,
		connected:      true,
		connectionType: UDPServerConnection,
		// We pass a function which can DELETE an open connection upon a disconnect.
		serverConnectionDeleter: func() {
			udpListener.openConnectionsLock.Lock()
			delete(udpListener.openConnections, newClientAddress)
			udpListener.openConnectionsLock.Unlock()
		},
	}
	udpListener.openConnectionsLock.RUnlock()

	return &newConnection, nil
}
