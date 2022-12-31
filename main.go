package main

import "MQTT-GO/gobro"

func main() {

	server := gobro.CreateServer()
	server.StartServer()

}
