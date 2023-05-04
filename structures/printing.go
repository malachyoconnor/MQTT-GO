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

// PrintInterface prints an interface in a nice format.
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

// PrintCentrally prints a string in the center of the terminal.
// It uses the Println function so it is thread safe.
func PrintCentrally(toPrint ...any) {
	return
	output := fmt.Sprint(toPrint...)
	padding := (int(terminalWidth) - len(output)) / 2
	fmt.Println(strings.Repeat(" ", padding) + output)
	return
	Println(strings.Repeat(" ", padding) + output)
	go func() {
	}()
}

// PrintItems prints the items in the linked list.
func (ll *LinkedList[T]) PrintItems() {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	fmt.Print("[ ")
	node := ll.head
	for i := 0; i < ll.Size(); i++ {
		fmt.Print(node.val, " ")
		node = node.next
	}
	fmt.Print("]")
}

// PrintArray prints an array in a nice format.
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

// Println is a thread safe version of fmt.Println.
func Println(a ...any) (int, error) {
	return 0, nil
	// printingMutex.Lock()
	// defer printingMutex.Unlock()
	return fmt.Println(a...)
}

// Printf is a thread safe version of fmt.Printf.
func Printf(format string, a ...any) (int, error) {
	printingMutex.Lock()
	defer printingMutex.Unlock()
	return fmt.Printf(format, a...)
}

// Print is a thread safe version of fmt.Print.
func Print(a ...any) (int, error) {
	printingMutex.Lock()
	defer printingMutex.Unlock()
	return fmt.Print(a...)
}
