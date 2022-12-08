package client

import "net"

type ClientID string

type ClientTable map[ClientID]*Client

type Client struct {
	// Should be a max of 23 characters!
	ClientIdentifier ClientID
	IPAddress        [4]byte
	Topics           []string
	TCPConnection    net.Conn
}
