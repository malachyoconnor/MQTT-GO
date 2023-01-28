package clients

import (
	"errors"
	"log"
	"strings"

	"MQTT-GO/structures"
)

// Rename to TopicToSubscriberStore
type TopicToSubscribers struct {
	topLevelMap *structures.SafeMap[string, *topic]
}

func (topicToSubs TopicToSubscribers) DeleteAll() {
	for _, topic := range topicToSubs.topLevelMap.Values() {
		topic.deleteSelf()
	}
}

func CreateTopicMap() *TopicToSubscribers {
	topicMap := TopicToSubscribers{
		topLevelMap: structures.CreateSafeMap[string, *topic](),
	}
	return &topicMap
}

func (topicMap *TopicToSubscribers) PrintTopics() {
	for _, topic := range topicMap.topLevelMap.Values() {
		topic.PrintTopics()
		ServerPrintln()
	}
}

func (topicMap *TopicToSubscribers) DeleteClientSubscriptions(client *Client) {
	clientTopics := client.Topics
	if clientTopics == nil {
		return
	}
	node := clientTopics.Head()
	for node != nil {
		topic := node.Value().TopicFilter
		clientLL, err := topicMap.get(topic)
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
		if clientLL.Size == 0 {
			err := topicMap.Delete(topic)
			if err != nil {
				panic(err)
			}
		}

		node = node.Next()
	}
}

func (topicMap *TopicToSubscribers) Put(topicName string, clientID ClientID) error {
	clientLL, err := topicMap.get(topicName)
	if err != nil {
		if err == ErrTopicDoesntExist {
			err := topicMap.AddTopic(topicName)
			if err != nil {
				return err
			}
			clientLL, _ = topicMap.get(topicName)
		} else {
			return err
		}
	}
	clientLL.Append(clientID)
	return nil
}

func (topicMap *TopicToSubscribers) PutClients(topicName string, clientIDs []ClientID) error {
	clientLL, err := topicMap.get(topicName)
	if err != nil {
		return err
	}
	for _, clientID := range clientIDs {
		clientLL.Append(clientID)
	}
	return nil
}

func (topicMap *TopicToSubscribers) Contains(topicName string) bool {
	_, err := topicMap.get(topicName)
	return err == nil
}

func (topicMap *TopicToSubscribers) Unsubscribe(clientID ClientID, topicNames ...string) {
	for _, topic := range topicNames {
		subscribedClients, err := topicMap.get(topic)
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
		if subscribedClients.Size == 0 {
			err := topicMap.Delete(topic)
			if err != nil {
				log.Println("- error while removing topic from topicMap", err)
				ServerPrintln("error while removing topic from topicMap", err)
			}
		}
	}
}

// Delete() deletes a topic from the topic map - it can return an ErrTopicDoesntExist error
// or nil
func (t *TopicToSubscribers) Delete(topicName string) error {
	topicSections := strings.Split(topicName, "/")

	if !t.topLevelMap.Contains(topicSections[0]) {
		return ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		t.topLevelMap.Delete(topicSections[0])
		return nil
	}

	val := t.topLevelMap.Get(topicSections[0])
	return val.DeleteTopic(topicSections[1:])
}

var ErrTopicAlreadyExists = errors.New("error: Trying to add client that already exists")

func (topicClientStore *TopicToSubscribers) AddTopic(topicName string) error {
	topicSections := strings.Split(topicName, "/")
	// If this is just a top level topic like sensors/ as opposed to sensors/c02sensors/...
	if len(topicSections) == 1 {
		if topicClientStore.topLevelMap.Contains(topicSections[0]) {
			return ErrTopicAlreadyExists
		}
		topicClientStore.topLevelMap.Put(topicSections[0], makeBaseTopic(topicSections[0]))
		return nil
	}

	if topicClientStore.topLevelMap.Contains(topicSections[0]) {
		topLevelTopic := topicClientStore.topLevelMap.Get(topicSections[0])
		return topLevelTopic.AddTopic(topicSections[1:])
	}
	baseTopic := makeBaseTopic(topicSections[0])
	topicClientStore.topLevelMap.Put(topicSections[0], baseTopic)
	return baseTopic.AddTopic(topicSections[1:])
}

func (topicToClient *TopicToSubscribers) GetMatchingClients(topicName string) (*structures.LinkedList[ClientID], error) {
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

	topLevelMap := topicToClient.topLevelMap
	topLevelTopic := topLevelMap.Get(topicSections[0])
	if topLevelTopic == nil {
		return nil, ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		return topLevelTopic.connectedClients, nil
	}
	result := topLevelTopic.getMatchingClients(topicSections[1:])
	if result == nil {
		return nil, ErrTopicDoesntExist
	}
	// We don't want to send a client the same message twice
	result.RemoveDuplicates()

	return result, nil
}

// err can be ErrTopicDoesntExist or nil
func (topicToClient *TopicToSubscribers) get(topicName string) (*structures.LinkedList[ClientID], error) {
	topicSections := strings.Split(topicName, "/")

	topLevelTopic := topicToClient.topLevelMap.Get(topicSections[0])
	if topLevelTopic == nil {
		return nil, ErrTopicDoesntExist
	}

	if len(topicSections) == 1 {
		return topLevelTopic.connectedClients, nil
	}

	result := topLevelTopic.get(topicSections[1:])
	if result == nil {
		return nil, ErrTopicDoesntExist
	}
	return result, nil
}

func (t *topic) PrintTopics() {
	ServerPrintf("%v (%v):  ", t.name, t.connectedClients.GetItems())
	for _, child := range t.children {
		ServerPrintf("%v ", child.name)
	}
	ServerPrintln()

	for _, child := range t.children {
		child.PrintTopics()
	}
}

type topic struct {
	name             string
	children         []*topic
	connectedClients *structures.LinkedList[ClientID]
}

func makeBaseTopic(topicName string) *topic {
	connectedClients := structures.CreateLinkedList[ClientID]()
	newTopic := topic{
		name:             topicName,
		children:         make([]*topic, 0, 5),
		connectedClients: connectedClients,
	}
	return &newTopic
}

var ErrTopicDoesntExist = errors.New("error: Topic doesn't exist")

func (t *topic) DeleteTopic(topicSections []string) error {
	if len(topicSections) == 1 {

		for i, child := range t.children {
			if child.name == topicSections[0] {
				child.deleteSelf()
				// Remove that child from your children
				t.children[i] = t.children[len(t.children)-1]
				t.children = t.children[:len(t.children)-1]
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

func (t *topic) deleteSelf() {
	if t == nil {
		return
	}
	for _, child := range t.children {
		child.deleteSelf()
	}
	t.connectedClients.DeleteLinkedList()
	t.connectedClients = nil
}

func (t *topic) AddTopic(topicSections []string) error {
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

func (t *topic) getAllLowerLevelClients() *structures.LinkedList[ClientID] {
	result := t.connectedClients
	for _, child := range t.children {
		result = structures.Concatenate(result, child.getAllLowerLevelClients())
	}
	return result
}

func (t *topic) getMatchingClients(topicSections []string) *structures.LinkedList[ClientID] {
	// If we've gotten to the end of the topic list
	if len(topicSections) == 0 {
		return t.connectedClients
	}

	if topicSections[0] == "#" {
		return t.getAllLowerLevelClients()
	}

	var result *structures.LinkedList[ClientID]

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
func (t *topic) get(topicSections []string) *structures.LinkedList[ClientID] {
	if len(topicSections) == 0 {
		return t.connectedClients
	}
	for _, child := range t.children {
		if child.name == topicSections[0] {
			return child.get(topicSections[1:])
		}
	}
	return nil
}
