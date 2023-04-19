package structures

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

var (
	terminalWidth = getTerminalWidth()
)

func PrintInterface(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	Println(string(s))
}

func getTerminalWidth() uint {
	width, err := terminal.Width()
	if err != nil {
		fmt.Println("Error getting terminal width:", err)
		return 300
	}
	return width
}

func PrintCentrally(toPrint ...any) {
	go func() {
		output := fmt.Sprint(toPrint...)
		padding := (int(terminalWidth) - len(output)) / 2
		Println(strings.Repeat(" ", padding) + output)
	}()
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
