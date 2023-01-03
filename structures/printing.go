package structures

import (
	"encoding/json"
	"fmt"
)

func PrintInterface(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

func center(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s))
}

func PrintCentrally(toPrint string) {
	fmt.Printf(center(toPrint, 200))
}
