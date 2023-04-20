package client

import (
	"fmt"
	"sync/atomic"
)

var (
	clientNumber = atomic.Int32{}
)

// FIXME:
// Previously, this was generating the same client names when stresstesting
// with multiple clients.
// For now it's been replaced with a simple number.
func generateRandomClientID() string {
	return fmt.Sprint(clientNumber.Add(1))
}
