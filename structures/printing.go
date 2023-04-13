package structures

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

func PrintInterface(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	Println(string(s))
}

func center(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s))
}

func PrintCentrally(toPrint ...string) {
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("  --")
	for _, str := range toPrint {
		stringBuilder.WriteString(str)
	}
	stringBuilder.WriteString("--")
	Println(stringBuilder.String())

	// Println(center(stringBuilder.String(), 100))
}

func (ll *LinkedList[T]) PrintItems() {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	fmt.Print("[ ")
	node := ll.head
	for i := 0; i < ll.Size; i++ {
		fmt.Print(node.val, " ")
		node = node.next
	}
	fmt.Print("]")
}

func PrintArray[T comparable](arr []T, defaultValue T) {
	fmt.Print("[ ")

	for _, val := range arr {
		if val != defaultValue {
			fmt.Print(val, " ")
		}
	}

	fmt.Print("]\n")
}

var (
	printingMutex = sync.Mutex{}
)

func Println(a ...any) (int, error) {
	printingMutex.Lock()
	defer printingMutex.Unlock()
	return fmt.Println(a...)
}

func Printf(format string, a ...any) (int, error) {
	printingMutex.Lock()
	defer printingMutex.Unlock()
	return fmt.Printf(format, a...)
}

func Print(a ...any) (int, error) {
	printingMutex.Lock()
	defer printingMutex.Unlock()
	return fmt.Print(a...)
}
