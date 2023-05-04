package clients

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"MQTT-GO/structures"
)

// Rename to TopicToSubscriberStore
type TopicTrie struct {
	topLevelMap *structures.SafeMap[string, *topicNode]
}

func (topicTrie TopicTrie) DeleteAll() {
	for _, topic := range topicTrie.topLevelMap.Values() {
		fmt.Println("DELETING TOPIC", topic.name)
		topic.deleteSelf()
	}
}

func CreateTopicTrie() *TopicTrie {
	topicTrie := TopicTrie{
		topLevelMap: structures.CreateSafeMap[string, *topicNode](),
	}
	return &topicTrie
}

func (topicTrie *TopicTrie) PrintTopics() {
	for _, topic := range topicTrie.topLevelMap.Values() {
		topic.PrintTopics()
		structures.Println()
	}
}

func (topicTrie *TopicTrie) DeleteClientSubscriptions(client *Client) {
	clientTopics := client.Topics
	if clientTopics == nil {
		return
	}
	node := clientTopics.Head()
	for node != nil {
		topic := node.Value().TopicFilter
		clientLL, err := topicTrie.get(topic)
		// Race condition
		if err != nil {
			ServerPrintln("Tried to remove topic that had already been deleted")
			node = node.Next()
			continue
		}

		err = clientLL.Delete(client.ClientIdentifier)
		if err != nil {
			ServerPrintln("Tried to delete client and got:", err)
		}
		if clientLL.Size() == 0 {
			err := topicTrie.Delete(topic)
			if err != nil {
				panic(err)
			}
		}

		node = node.Next()
	}
}

func (topicTrie *TopicTrie) Put(topicName string, clientID ClientID) error {
	clientLL, err := topicTrie.get(topicName)
	if err != nil {
		if errors.Is(err, ErrTopicDoesntExist) {
			err = topicTrie.AddTopic(topicName)
			if err != nil {
				return err
			}
			clientLL, _ = topicTrie.get(topicName)
		} else {
			return err
		}
	}
	clientLL.Append(clientID)
	return nil
}

func (topicTrie *TopicTrie) PutClients(topicName string, clientIDs []ClientID) error {
	clientLL, err := topicTrie.get(topicName)
	if err != nil {
		return err
	}
	for _, clientID := range clientIDs {
		clientLL.Append(clientID)
	}
	return nil
}

func (topicTrie *TopicTrie) Contains(topicName string) bool {
	_, err := topicTrie.get(topicName)
	return err == nil
}

func (topicTrie *TopicTrie) Unsubscribe(clientID ClientID, topicNames ...string) {
	for _, topic := range topicNames {
		subscribedClients, err := topicTrie.get(topic)
		if err != nil {
			ServerPrintln("Error while unsubscribing:", err)
			continue
		}
		err = subscribedClients.Delete(clientID)
		if err != nil {
			ServerPrintln("Error while deleting client:", err)
			continue
		}
		// If no one is left subscribed to the topic, remove it.
		// This is to avoid memory leaks
		if subscribedClients.Size() == 0 {
			err := topicTrie.Delete(topic)
			if err != nil {
				log.Println("- error while removing topic from topicTrie", err)
				ServerPrintln("error while removing topic from topicTrie", err)
			}
		}
	}
}

// Delete() deletes a topic from the topic map - it can return an ErrTopicDoesntExist error
// or nil
func (topicTrie *TopicTrie) Delete(topicName string) error {
	topicSections := strings.Split(topicName, "/")

	if !topicTrie.topLevelMap.Contains(topicSections[0]) {
		return ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		topicTrie.topLevelMap.Delete(topicSections[0])
		return nil
	}

	val := topicTrie.topLevelMap.Get(topicSections[0])
	return val.DeleteTopic(topicSections[1:])
}

var ErrTopicAlreadyExists = errors.New("error: Trying to add client that already exists")

func (topicTrie *TopicTrie) AddTopic(topicName string) error {
	topicSections := strings.Split(topicName, "/")
	// If this is just a top level topic like sensors/ as opposed to sensors/c02sensors/...
	if len(topicSections) == 1 {
		if topicTrie.topLevelMap.Contains(topicSections[0]) {
			return ErrTopicAlreadyExists
		}
		topicTrie.topLevelMap.Put(topicSections[0], makeBaseTopic(topicSections[0]))
		return nil
	}

	if topicTrie.topLevelMap.Contains(topicSections[0]) {
		topLevelTopic := topicTrie.topLevelMap.Get(topicSections[0])
		return topLevelTopic.AddTopic(topicSections[1:])
	}
	baseTopic := makeBaseTopic(topicSections[0])
	topicTrie.topLevelMap.Put(topicSections[0], baseTopic)
	return baseTopic.AddTopic(topicSections[1:])
}

func (topicTrie *TopicTrie) GetMatchingClients(topicName string) (*structures.LinkedList[ClientID], error) {
	if topicName == "" {
		return nil, errors.New("error: Cannot search for empty topic")
	}

	if topicName[len(topicName)-1] == '/' {
		return nil, errors.New("error: Wildcard topics cannot end with /")
	}
	topicSections := strings.Split(topicName, "/")

	if len(topicSections) > 1 {
		for _, topic := range topicSections[:len(topicSections)-2] {
			if topic == "#" {
				return nil, errors.New("error: # wildcard should only be at the end of a topic subscription")
			}
		}
	}

	if topicSections[0] == "+" {
		result := structures.CreateLinkedList[ClientID]()
		for _, topLevelTopic := range topicTrie.topLevelMap.Values() {
			result = structures.Concatenate(result, topLevelTopic.getMatchingClients(topicSections[1:]))
		}
		result.RemoveDuplicates()
		return result, nil
	}

	topLevelMap := topicTrie.topLevelMap
	topLevelTopic := topLevelMap.Get(topicSections[0])
	if topLevelTopic == nil {
		return nil, ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		return topLevelTopic.subscribedClients.DeepCopy(), nil
	}
	result := topLevelTopic.getMatchingClients(topicSections[1:])
	if result.Size() == 0 {
		return nil, ErrTopicDoesntExist
	}
	// We don't want to send a client the same message twice
	result.RemoveDuplicates()

	return result, nil
}

// err can be ErrTopicDoesntExist or nil
func (topicTrie *TopicTrie) get(topicName string) (*structures.LinkedList[ClientID], error) {
	topicSections := strings.Split(topicName, "/")

	topLevelTopic := topicTrie.topLevelMap.Get(topicSections[0])
	if topLevelTopic == nil {
		return nil, ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		return topLevelTopic.subscribedClients, nil
	}

	result := topLevelTopic.get(topicSections[1:])
	if result == nil {
		return nil, ErrTopicDoesntExist
	}
	return result, nil
}

func (t *topicNode) PrintTopics() {
	ServerPrintf("%v (%v):  ", t.name, t.subscribedClients.GetItems())
	for _, child := range t.children {
		ServerPrintf("%v ", child.name)
	}
	structures.Println()

	for _, child := range t.children {
		child.PrintTopics()
	}
}

type topicNode struct {
	name              string
	children          []*topicNode
	subscribedClients *structures.LinkedList[ClientID]
}

func makeBaseTopic(topicName string) *topicNode {
	connectedClients := structures.CreateLinkedList[ClientID]()
	newTopic := topicNode{
		name:              topicName,
		children:          make([]*topicNode, 0, 5),
		subscribedClients: connectedClients,
	}
	return &newTopic
}

var ErrTopicDoesntExist = errors.New("error: Topic doesn't exist")

func (t *topicNode) DeleteTopic(topicSections []string) error {
	if len(topicSections) == 1 {
		for i, child := range t.children {
			if child.name == topicSections[0] {
				child.deleteSelf()
				// Remove that child from your children
				t.children = append(t.children[:i], t.children[i+1:]...)
				return nil
			}
		}
		return ErrTopicDoesntExist
	}

	for _, child := range t.children {
		if child.name == topicSections[0] {
			return child.DeleteTopic(topicSections[1:])
		}
	}

	return ErrTopicDoesntExist
}

func (t *topicNode) deleteSelf() {
	if t == nil {
		return
	}
	for _, child := range t.children {
		child.deleteSelf()
	}
	fmt.Println("DELETING LINKED LIST")
	t.subscribedClients.DeleteLinkedList()
	t.subscribedClients = nil
}

func (t *topicNode) AddTopic(topicSections []string) error {
	// If the length of topicSections is one we just add it to our children
	if len(topicSections) == 1 {
		for _, child := range t.children {
			if child.name == topicSections[0] {
				return ErrTopicAlreadyExists
			}
		}
		t.children = append(t.children, makeBaseTopic(topicSections[0]))
		return nil
	}
	// We follow an existing chain of subtopics down as far as we can go
	for _, child := range t.children {
		if child.name == topicSections[0] {
			return child.AddTopic(topicSections[1:])
		}
	}
	// If we can't find any matching sub topics
	resultChain := makeBaseTopic(topicSections[0])
	t.children = append(t.children, resultChain)

	for _, topicSection := range topicSections[1:] {
		newTopic := makeBaseTopic(topicSection)
		resultChain.children = append(resultChain.children, newTopic)
		resultChain = newTopic
	}
	return nil
}

func (t *topicNode) getAllLowerLevelClients() *structures.LinkedList[ClientID] {
	result := t.subscribedClients
	for _, child := range t.children {
		result = structures.Concatenate(result, child.getAllLowerLevelClients())
	}
	return result
}

func (t *topicNode) getMatchingClients(topicSections []string) *structures.LinkedList[ClientID] {
	// If we've gotten to the end of the topic list
	if len(topicSections) == 0 {
		return t.subscribedClients
	}

	if topicSections[0] == "#" {
		return t.getAllLowerLevelClients()
	}

	result := structures.CreateLinkedList[ClientID]()

	// These two if statements allow for publishing to a wildcard!
	if len(topicSections) == 1 && t.name == topicSections[0] {
		return t.subscribedClients
	}

	if topicSections[0] == "+" {
		for _, child := range t.children {
			result = structures.Concatenate(result, child.getMatchingClients(topicSections[1:]))
		}
		return result
	}

	for _, child := range t.children {
		// If we're not at the bottom level topic
		if child.name == topicSections[0] || child.name == "+" {
			result = structures.Concatenate(result, child.getMatchingClients(topicSections[1:]))
		} else if child.name == "#" {
			result = structures.Concatenate(result, child.getAllLowerLevelClients())
		}
	}
	return result
}

// Can return a client list of nil (if the topic doesn't exist)
func (t *topicNode) get(topicSections []string) *structures.LinkedList[ClientID] {
	if len(topicSections) == 0 {
		return t.subscribedClients
	}
	for _, child := range t.children {
		if child.name == topicSections[0] {
			return child.get(topicSections[1:])
		}
	}
	return nil
}
