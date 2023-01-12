package packets

import (
	"fmt"
)

func PrintBinary(toPrint []byte) {
	for _, b := range toPrint {
		fmt.Printf("%08b\n", b)
	}
}
