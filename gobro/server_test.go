package gobro_test

import (
	"MQTT-GO/gobro"
	"testing"
	"time"
)

func TestServerStarts(t *testing.T) {

	defer func() {
		err := recover()
		if err != nil {
			t.Error("Server crashed", err)
		}

	}()

	server := gobro.CreateServer()
	go server.StartServer()

	time.Sleep(time.Millisecond * 200)

}
