package client

import (
	"fmt"
	"net"
)

// Here we'll have the functions that make the client perform it's actions.

func ConnectToServer(ip string, port int) {
	address := net.JoinHostPort(ip, fmt.Sprint(port))

	connection, err := net.Dial("tcp", address)

	if err != nil {
		fmt.Println(err)
		return
	}

	connection.Write([]byte{123})
}
