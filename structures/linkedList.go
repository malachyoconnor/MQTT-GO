package structures

import (
	"errors"
	"fmt"
	"sync"
)

// LinkedList is a thread safe linked list implementation.
type LinkedList[T comparable] struct {
	head *Node[T]
	tail *Node[T]
	size int
	lock sync.RWMutex
}

// CreateLinkedList creates a new linked list.
func CreateLinkedList[T comparable]() *LinkedList[T] {
	result := LinkedList[T]{
		head: nil,
		tail: nil,
		lock: sync.RWMutex{},
		size: 0,
	}
	return &result
}

// GetItems returns a slice of the items in the linked list.
func (ll *LinkedList[T]) GetItems() []T {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	result := make([]T, 0, ll.size)

	node := ll.head
	for i := 0; i < ll.size; i++ {
		result = append(result, node.val)
		node = node.next
	}

	return result
}

// DeleteLinkedList deletes the linked list, setting all pointers to nil.
// This is useful for garbage collection.
func (ll *LinkedList[T]) DeleteLinkedList() {
	if ll == nil {
		return
	}
	ll.lock.Lock()
	defer ll.lock.Unlock()

	node := ll.head
	for node != nil {
		nextNode := node.next
		node.next = nil
		node = nextNode
	}
}

// Concatenate takes two linked lists and returns a new linked list that is the concatenation of the two.
// This concatenation is slow as we combine the two lists we DON'T copy them
// If this is really slow then we can change that and just make sure we fix the lists when we're done
// The point is that holding onto either of the other lists could be bad for concurrency - instead
// we take a snapshot.
func Concatenate[T comparable](llA *LinkedList[T], llB *LinkedList[T]) *LinkedList[T] {
	result := CreateLinkedList[T]()

	if llA == nil && llB == nil {
		return nil
	}
	if llA == nil {
		return llB.DeepCopy()
	}
	if llB == nil {
		return llA.DeepCopy()
	}

	llA.lock.RLock()

	nodeA := llA.head
	for i := 0; i < llA.size; i++ {
		result.Append(nodeA.val)
		nodeA = nodeA.next
	}
	llA.lock.RUnlock()

	llB.lock.RLock()
	nodeB := llB.head
	for i := 0; i < llB.size; i++ {
		result.Append(nodeB.val)
		nodeB = nodeB.next
	}
	llB.lock.RUnlock()

	return result
}

// DeepCopy returns a deep copy of the linked list.
func (ll *LinkedList[T]) DeepCopy() *LinkedList[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	result := CreateLinkedList[T]()

	if ll.size == 0 {
		return result
	}

	result.size = ll.size
	node := ll.head
	result.head = &Node[T]{val: node.val}
	resultNode := result.head

	for i := 0; i < ll.size-1; i++ {
		node = node.next
		resultNode.next = &Node[T]{val: node.val, prev: resultNode}
		resultNode = resultNode.next
	}

	result.tail = resultNode

	return result
}

// CombineLinkedLists takes a variable number of linked lists and returns a new linked list that is the concatenation of all
// the lists.
func CombineLinkedLists[T comparable](lists ...*LinkedList[T]) *LinkedList[T] {
	result := CreateLinkedList[T]()
	for _, list := range lists {
		list.lock.RLock()
		defer list.lock.RUnlock()
	}

	for _, list := range lists {
		node := list.Head()
		for i := 0; i < list.size; i++ {
			result.Append(node.val)
			node = node.next
		}
	}

	return result
}

// Head returns the head of the linked list.
func (ll *LinkedList[T]) Head() *Node[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	return ll.head
}

// Contains returns true if the linked list contains the given value.
func (ll *LinkedList[T]) Contains(val T) bool {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	node := ll.head
	for i := 0; i < ll.size; i++ {
		if node.val == val {
			return true
		}
		node = node.next
	}
	return false
}

// RemoveDuplicates removes all duplicates from the linked list.
func (ll *LinkedList[T]) RemoveDuplicates() {
	ll.lock.Lock()
	defer ll.lock.Unlock()
	existingItems := make(map[T]bool, ll.size)

	node := ll.head
	for node != nil {
		_, found := existingItems[node.val]
		if !found {
			existingItems[node.val] = true
		} else {
			switch node {
			case ll.head:
				{
					ll.head = ll.head.next
					ll.head.prev = nil
				}
			case ll.tail:
				ll.tail = ll.tail.prev
				ll.tail.next = nil
			default:
				node.deleteNode()
			}
			ll.size--
		}
		node = node.next
	}
}

// Append appends a value to the end of the linked list.
func (ll *LinkedList[T]) Append(val T) {
	ll.lock.Lock()
	defer ll.lock.Unlock()
	// If the list is empty
	newNode := &Node[T]{val: val}
	ll.size++

	switch {
	case ll.head == nil:
		{
			ll.head = newNode
			ll.tail = newNode
		}
	case ll.tail == nil:
		{
			panic("error: Head of linked list is not nil and the tail is")
		}
	default:
		{
			ll.tail.next = newNode
			newNode.prev = ll.tail
			ll.tail = newNode
		}
	}
}

// Delete deletes a value from the linked list.
func (ll *LinkedList[T]) Delete(val T) error {
	errValDoesntExist := errors.New(fmt.Sprint("error: Value", val, "not found in linked list"))

	ll.lock.Lock()
	defer ll.lock.Unlock()

	// If we've removed an item and left one item in the list - we need to ensure that
	// we don't get infinite loops when we go to search our list.
	defer func(linkedList *LinkedList[T]) {
		if linkedList.size == 1 {
			ll.head.next = nil
			ll.tail.prev = nil
			ll.tail = nil
		}
	}(ll)

	// If the value we're deleting is the head or the tail
	// then we need to adjust the linked list's head/tail

	if ll.head == nil {
		return errValDoesntExist
	}

	// If we've only got one item in the list
	if ll.tail == ll.head && ll.head.val == val {
		ll.head, ll.tail = nil, nil
		ll.size--
		return nil
	}

	if ll.head.val == val {
		ll.head = ll.head.next
		ll.size--
		return nil
	} else if ll.tail.val == val {
		ll.tail = ll.tail.prev
		ll.size--
		return nil
	}

	node := ll.head
	for i := 0; i < ll.size; i++ {
		if node.val == val {
			ll.size--
			node.deleteNode()
			return nil
		}
		node = node.next
	}

	return errValDoesntExist
}

// ReadLock locks the linked list for reading.
func (ll *LinkedList[T]) Size() int {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	return ll.size
}

// Filter returns a new linked list that contains only the items that match the filter.
func (ll *LinkedList[T]) Filter(filter func(T) bool) *LinkedList[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	result := CreateLinkedList[T]()
	node := ll.head
	for i := 0; i < ll.size; i++ {
		if filter(node.val) {
			result.Append(node.val)
		}
		node = node.next
	}
	return result
}

// FilterSingleItem returns a single item that matches the filter.
// Note this returns a POINTER to the item - because it needed to be able to return nil if naught was found
func (ll *LinkedList[T]) FilterSingleItem(filter func(T) bool) *T {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	node := ll.head
	for i := 0; i < ll.size; i++ {
		if filter(node.val) {
			return &node.val
		}
		node = node.next
	}
	return nil
}

// Node is a node in a linked list. It contains a pointer to the next node and the previous node
// as well as the value of the node.
type Node[T comparable] struct {
	prev *(Node[T])
	next *(Node[T])
	val  T
}

// Value returns the value of the node.
func (node *Node[T]) Value() T {
	return node.val
}

// Next returns the next node in the linked list.
func (node *Node[T]) Next() *Node[T] {
	return node.next
}

// Prev returns the previous node in the linked list.
func (node *Node[T]) Prev() *Node[T] {
	return node.prev
}

func (node *Node[T]) deleteNode() {
	// Note prev CANNOT be nil, as we cannot be called on the head
	// We can't rely on that because we can't delete ourselves if we're
	// the head or tail because we won't be garbage collected.
	node.prev.next = node.next
	node.next.prev = node.prev
}
