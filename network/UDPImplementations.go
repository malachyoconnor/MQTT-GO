package network

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

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

	conn.connectionType = UDP_CLIENT_CONNECTION

	go func() {
		for {
			buffer := make([]byte, 1024)
			bytesRead, receivedAddr, err := conn.connection.ReadFromUDP(buffer)
			buffer = buffer[:bytesRead]

			if errors.Is(err, net.ErrClosed) {
				fmt.Println("Connection closed, disconnecting")
				return
			}

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
	}()
	conn.connected = true

	return err
}

func (conn *UDPCon) Write(toWrite []byte) (n int, err error) {
	// If we're a client, we've created a singly connected connection.
	// If we're a server, we've got a general purpose connection with which we can
	// send to multiple addresses.
	if conn.connectionType == UDP_SERVER_CONNECTION {
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

func (conn *UDPCon) Close() error {
	// We don't want to close the connection on the other end
	// So we just send a disconnect and stop listening.

	if conn.connectionType == UDP_CLIENT_CONNECTION {
		fmt.Println("Closing a connection from:", conn.localAddr, "to", conn.remoteAddr)
		return (*conn.connection).Close()
	}

	if conn.connectionType == UDP_SERVER_CONNECTION {
		conn.serverConnectionDeleter()
		close(conn.packetBuffer)
	}
	return nil

}

func (conn *UDPCon) RemoteAddr() net.Addr {
	remoteAddr, err := net.ResolveUDPAddr("udp", conn.remoteAddr)
	if err != nil {
		fmt.Println("Error while resolving UDP address:", err)
		return nil
	}
	return remoteAddr
}

func (conn *UDPCon) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

func (conn *UDPCon) SetDeadline(t time.Time) error {
	return nil
}

func (conn *UDPCon) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *UDPCon) SetWriteDeadline(t time.Time) error {
	return nil
}

// Next the listening methods

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

func startUDPbackgroundListener(udpListener *UDPListener, connection *net.UDPConn) {
	for {
		buffer := make([]byte, 1024)
		bytesRead, receivedAddr, err := connection.ReadFromUDP(buffer)
		if errors.Is(err, net.ErrClosed) {
			fmt.Println("Connection closed, ceasing to listen")
			return
		}
		if err != nil {
			fmt.Println(err)
		}
		buffer = buffer[:bytesRead]
		address := fmt.Sprint(receivedAddr.IP.String(), ":", receivedAddr.Port)
		udpListener.openConnectionsLock.RLock()
		if _, found := udpListener.openConnections[address]; !found {

			if udpListener.numClientsToAdd.Load() > 0 {
				udpListener.openConnectionsLock.RUnlock()
				udpListener.openConnectionsLock.Lock()
				udpListener.openConnections[address] = make(chan []byte, 512)
				udpListener.openConnections[address] <- buffer
				udpListener.openConnectionsLock.Unlock()
				udpListener.newClientBuffer <- address
				udpListener.numClientsToAdd.Add(-1)
				continue
			}

			fmt.Println("Received a message from an unconnected address:", address, "message:", buffer)
			continue
		}
		udpListener.openConnections[address] <- buffer
		udpListener.openConnectionsLock.RUnlock()
	}
}

func (udpListener *UDPListener) Close() error {
	return udpListener.listener.Close()
}

// Note that UDP is connectionless, so we need to create our own connections. That means listening for ALL UDP packets
// and pushing them down the correct connection.

func (udpListener *UDPListener) Accept() (net.Conn, error) {
	udpListener.numClientsToAdd.Add(1)
	newClientAddress := <-udpListener.newClientBuffer

	udpListener.openConnectionsLock.RLock()

	newConnection := UDPCon{
		packetBuffer:   udpListener.openConnections[newClientAddress],
		remoteAddr:     newClientAddress,
		localAddr:      udpListener.listener.LocalAddr().String(),
		connection:     udpListener.listener,
		connected:      true,
		connectionType: UDP_SERVER_CONNECTION,
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
