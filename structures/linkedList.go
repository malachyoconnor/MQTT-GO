package structures

import (
	"errors"
	"fmt"
	"sync"
)

type LinkedList[T comparable] struct {
	head *Node[T]
	tail *Node[T]
	Size int
	lock sync.RWMutex
}

func CreateLinkedList[T comparable]() LinkedList[T] {
	return LinkedList[T]{
		head: nil,
		tail: nil,
		Size: 0,
	}
}

// This concatenation is slow as we combine the two lists we DON'T copy them
// If this is really slow then we can change that and just make sure we fix the lists when we're done
// The point is that holding onto either of the other lists could be bad for concurrency - instead
// we take a snapshot.
func Concatenate[T comparable](llA *LinkedList[T], llB *LinkedList[T]) *LinkedList[T] {

	result := CreateLinkedList[T]()

	llA.lock.RLock()

	nodeA := llA.head
	for nodeA != nil {
		result.Append(nodeA.val)
		nodeA = nodeA.next
	}
	llA.lock.RUnlock()

	llB.lock.RLock()
	nodeB := llB.head
	for nodeB != nil {
		result.Append(nodeB.val)
		nodeB = nodeB.next
	}
	llB.lock.RUnlock()

	return &result
}

func CombineLinkedLists[T comparable](lists ...*LinkedList[T]) *LinkedList[T] {
	result := CreateLinkedList[T]()

	for _, list := range lists {
		list.lock.RLock()
		node := list.Head()
		for node != nil {
			result.Append(node.val)
			node = node.next
		}
		list.lock.RUnlock()
	}

	return &result
}

func (ll *LinkedList[T]) Head() *Node[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	return (*ll).head
}

func (ll *LinkedList[T]) Contains(val T) bool {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	node := ll.head
	for i := 0; i < ll.Size; i++ {
		if node.val == val {
			return true
		}
		node = node.next
	}
	return false
}

func (ll *LinkedList[T]) Append(val T) {
	ll.lock.Lock()
	defer ll.lock.Unlock()
	// If the list is empty
	newNode := &Node[T]{val: val}
	ll.Size += 1
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
	} else if ll.tail == nil {
		panic("error: Head of linked list is not nil and the tail is")
	} else {
		ll.tail.next = newNode
		(*newNode).prev = ll.tail
		ll.tail = newNode
	}
}

func (ll *LinkedList[T]) Delete(val T) error {
	ll.lock.Lock()
	defer ll.lock.Unlock()
	// If the value we're deleting is the head or the tail
	// then we need to adjust the linked list's head/tail
	if ll.head.val == val {
		ll.head = ll.head.next
		ll.Size -= 1
		return nil
	} else if ll.tail.val == val {
		ll.tail = ll.tail.prev
		ll.Size -= 1
		return nil
	}

	node := ll.head
	for node := node.next; node != nil; {
		if node.val == val {
			ll.Size -= 1
			node.Delete()
			return nil
		}
	}
	return errors.New(fmt.Sprint("error: Value", val, "not found in linked list"))
}

type Node[T comparable] struct {
	prev *(Node[T])
	next *(Node[T])
	val  T
}

func (node *Node[T]) Value() T {
	return node.val
}

func (node *Node[T]) Next() *Node[T] {
	return (*node).next
}

func (node *Node[T]) Prev() *Node[T] {
	return (*node).prev
}

func (node *Node[T]) Delete() {
	// If we're the head
	if node.prev == nil {
		if node.next != nil {
			// If we've got a next node - then set that as us
			node = node.next
		} else {
			// Otherwise we've got no next or previous item - so just delete the whole list
			node = nil
		}
	} else {
		node.prev.next = node.next
	}
}
