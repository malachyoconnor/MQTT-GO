package structures_test

import (
	"MQTT-GO/structures"
	"testing"

	"golang.org/x/exp/slices"
)

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
