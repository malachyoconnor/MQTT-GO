package structures

type Node struct {
	prev *Node
	next *Node
	val  interface{}
}

func (node *Node) Delete() {
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

// func (node *Node) append(newValue interface{}) {

// }
