package structures_test

import (
	"MQTT-GO/structures"
	"fmt"
	"testing"

	"golang.org/x/exp/slices"
)

func TestAddingItems(t *testing.T) {
	linkedList := structures.CreateLinkedList[int]()

	for i := 0; i < 10; i++ {
		linkedList.Append(i)
	}

	fmt.Println(linkedList.GetItems())

	if !slices.Equal(linkedList.GetItems(), []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Error("Remove duplicates not working correctly.")
	}
}

func TestRemovingItems(t *testing.T) {
	linkedList := structures.CreateLinkedList[int]()

	for i := 0; i < 10; i++ {
		linkedList.Append(i)
	}

	linkedList.Delete(2)
	linkedList.Delete(8)
	fmt.Println(linkedList.GetItems())

	if !slices.Equal(linkedList.GetItems(), []int{0, 1, 3, 4, 5, 6, 7, 9}) {
		t.Error("Remove duplicates not working correctly.")
	}
}

func TestRemovingDuplicates(t *testing.T) {
	linkedList := structures.CreateLinkedList[int]()
	linkedList.Append(1)
	linkedList.Append(2)
	linkedList.Append(2)
	linkedList.Append(3)

	linkedList.RemoveDuplicates()
	if !slices.Equal(linkedList.GetItems(), []int{1, 2, 3}) {
		t.Error("Remove duplicates not working correctly.")
	}
}
