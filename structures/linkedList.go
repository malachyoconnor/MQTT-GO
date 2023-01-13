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

func CreateLinkedList[T comparable]() *LinkedList[T] {
	result := LinkedList[T]{
		head: nil,
		tail: nil,
		Size: 0,
	}
	return &result
}

func (ll *LinkedList[T]) GetItems() []T {
	result := make([]T, 0, ll.Size)

	node := ll.head
	for node != nil {
		result = append(result, node.val)
		node = node.next
	}
	return result
}

func (ll *LinkedList[T]) DeleteLinkedList() {
	if ll == nil {
		return
	}
	node := ll.head
	for node != nil {
		nextNode := node.next
		node.next = nil
		node = nextNode
	}
}

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

	return result
}

func (ll *LinkedList[T]) DeepCopy() *LinkedList[T] {
	result := CreateLinkedList[T]()
	node := ll.head
	for node != nil {
		result.Append(node.val)
		node = node.next
	}
	return result
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

	return result
}

func (ll *LinkedList[T]) Head() *Node[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	return ll.head
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

func (ll *LinkedList[T]) RemoveDuplicates() {
	existingItems := make(map[T]bool, ll.Size)

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
				node.delete()
			}
			ll.Size--
		}
		node = node.next
	}
}

func (ll *LinkedList[T]) Append(val T) {
	ll.lock.Lock()
	defer ll.lock.Unlock()
	// If the list is empty
	newNode := &Node[T]{val: val}
	ll.Size++
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
	} else if ll.tail == nil {
		panic("error: Head of linked list is not nil and the tail is")
	} else {
		ll.tail.next = newNode
		newNode.prev = ll.tail
		ll.tail = newNode
	}
}

func (ll *LinkedList[T]) Delete(val T) error {
	ErrValDoesntExist := errors.New(fmt.Sprint("error: Value", val, "not found in linked list"))

	ll.lock.Lock()
	defer ll.lock.Unlock()

	// If we've removed an item and left one item in the list - we need to ensure that
	// we don't get infinite loops when we go to search our list.
	defer func() {
		if ll.Size == 1 {
			ll.head.next = nil
			ll.tail.prev = nil
		}
	}()

	// If the value we're deleting is the head or the tail
	// then we need to adjust the linked list's head/tail

	if ll.head == nil {
		return ErrValDoesntExist
	}

	// If we've only got one item in the list
	if ll.tail == ll.head && ll.head.val == val {
		ll.head, ll.tail = nil, nil
		ll.Size--
		return nil
	}

	if ll.head.val == val {
		ll.head = ll.head.next
		ll.Size--
		return nil
	} else if ll.tail.val == val {
		ll.tail = ll.tail.prev
		ll.Size--
		return nil
	}

	node := ll.head
	for i := 0; i < ll.Size; i++ {
		if node.val == val {
			ll.Size--
			node.delete()
			return nil
		}
		node = node.next
	}

	return ErrValDoesntExist
}

func (ll *LinkedList[T]) Filter(f func(T) bool) *LinkedList[T] {
	ll.lock.RLock()
	defer ll.lock.RUnlock()
	result := CreateLinkedList[T]()
	node := ll.head
	for node != nil {
		if f(node.val) {
			result.Append(node.val)
		}
		node = node.next
	}
	return result
}

// Note this returns a POINTER to the item - because it needed to be able to return nil if naught was found
func (ll *LinkedList[T]) FilterSingleItem(f func(T) bool) *T {
	ll.lock.RLock()
	defer ll.lock.RUnlock()

	node := ll.head
	for node != nil {
		if f(node.val) {
			return &node.val
		}
		node = node.next
	}
	return nil
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
	return node.next
}

func (node *Node[T]) Prev() *Node[T] {
	return node.prev
}

func (node *Node[T]) delete() {
	// Note prev CANNOT be nil, as we cannot be called on the head
	// We can't rely on that because we can't delete ourselves if we're
	// the head or tail because we won't be garbage collected.
	node.prev.next = node.next
	node.next.prev = node.prev
}
