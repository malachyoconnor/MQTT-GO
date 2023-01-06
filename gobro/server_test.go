package gobro

import (
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

	server := CreateServer()
	go server.StartServer()

	time.Sleep(time.Millisecond * 200)

}
