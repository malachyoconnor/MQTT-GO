package structures

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestRemovingDuplicates(t *testing.T) {

	ll := CreateLinkedList[int]()
	ll.Append(1)
	ll.Append(2)
	ll.Append(2)
	ll.Append(3)

	ll.RemoveDuplicates()
	if !slices.Equal(ll.GetItems(), []int{1, 2, 3}) {
		t.Error("Remove duplicates not working correctly.")
	}

}
