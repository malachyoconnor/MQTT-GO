package client

import (
	"fmt"
)

var (
	clientNumber = 0
)

// FIXME:
// Previously, this was generating the same client names when stresstesting
// with multiple clients.
// For now it's been replaced with a simple number.
func generateRandomClientID() string {
	clientNumber += 1
	return fmt.Sprint(clientNumber)

}

// 	stringBuilder := strings.Builder{}

// 	for i := 0; i < rand.Intn(5); i++ {
// 		stringBuilder.WriteByte('a' + byte(rand.Intn(25)))
// 	}

// 	return stringBuilder.String()
// }
