package structures

import (
	"errors"
	"fmt"
)

type LinkedList[T comparable] struct {
	head *Node[T]
	tail *Node[T]
	Size int
}

func CreateLinkedList[T comparable]() LinkedList[T] {
	return LinkedList[T]{
		head: nil,
		tail: nil,
		Size: 0,
	}
}

func (ll *LinkedList[T]) PrintItems() {
	node := ll.head
	for i := 0; i < ll.Size; i++ {
		fmt.Print(node.val, " ")
		node = node.next
	}

}

func (ll *LinkedList[T]) Append(val T) {
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
	if ll.head.val == val {
		ll.Size -= 1
		return nil
	}
	node := ll.head
	for node := node.next; node != nil; {
		if node.val == val {
			ll.Size -= 1
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
