package tools

import (
	h "encoding/hex"
	"fmt"
)

func Test() {
	fmt.Println("??")
}

func DecodeHex(hex string) {
	result, err := h.DecodeString(hex)

	for x := range result {
		fmt.Printf("%b\n", x)
	}

	fmt.Println(result, err)

	fmt.Println("Got here")
}
