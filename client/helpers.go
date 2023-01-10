package client

import (
	"math/rand"
	"strings"
	"time"
)

func generateRandomClientID() string {
	stringBuilder := strings.Builder{}
	rand.Seed(time.Now().Unix())

	for i := 0; i < rand.Intn(5); i++ {
		stringBuilder.WriteByte('a' + byte(rand.Intn(25)))
	}

	return stringBuilder.String()

}
