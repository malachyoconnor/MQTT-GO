package gobro_test

import (
	"testing"
	"time"

	"MQTT-GO/gobro"
)

func TestServerStarts(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Error("Server crashed", err)
		}
	}()

	server := gobro.NewServer()
	go server.StartServer("localhost", 8000)

	time.Sleep(time.Millisecond * 200)
	server.StopServer(false)
}
