package clients

import (
	"MQTT-GO/structures"
	"errors"
	"fmt"
	"strings"
)

type TopicsToClientStore struct {
	topLevelMap *structures.SafeMap[string, *topic]
}

func CreateTopicMap() *TopicsToClientStore {
	topicMap := TopicsToClientStore{
		topLevelMap: structures.CreateSafeMap[string, *topic](),
	}
	return &topicMap
}

func (topicMap *TopicsToClientStore) PrintTopics() {
	for _, topic := range topicMap.topLevelMap.Values() {
		topic.PrintTopics()
		fmt.Println()
	}
}

func (topicMap *TopicsToClientStore) Put(topicName string, clientID ClientID) error {

	clientLL, err := topicMap.get(topicName)
	if err != nil {
		if err == ErrTopicDoesntExist {
			topicMap.AddTopic(topicName)
			clientLL, _ = topicMap.get(topicName)
		} else {
			return err
		}
	}
	clientLL.Append(clientID)
	return nil
}

func (topicMap *TopicsToClientStore) PutClients(topicName string, clientIDs []ClientID) error {
	clientLL, err := topicMap.get(topicName)
	if err != nil {
		return err
	}
	for _, clientID := range clientIDs {
		clientLL.Append(clientID)
	}
	return nil
}

func (t *TopicsToClientStore) DeleteTopic(topicName string) error {
	topicSections := strings.Split(topicName, "/")

	if !t.topLevelMap.Exists(topicSections[0]) {
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

func (topicClientStore *TopicsToClientStore) AddTopic(topicName string) error {
	topicSections := strings.Split(topicName, "/")
	// If this is just a top level topic like sensors/ as opposed to sensors/c02sensors/...
	if len(topicSections) == 1 {
		if topicClientStore.topLevelMap.Exists(topicSections[0]) {
			return ErrTopicAlreadyExists
		}
		topicClientStore.topLevelMap.Put(topicSections[0], makeBaseTopic(topicSections[0]))
		return nil
	}

	if topicClientStore.topLevelMap.Exists(topicSections[0]) {
		topLevelTopic := topicClientStore.topLevelMap.Get(topicSections[0])
		return topLevelTopic.AddTopic(topicSections[1:])
	} else {
		baseTopic := makeBaseTopic(topicSections[0])
		topicClientStore.topLevelMap.Put(topicSections[0], baseTopic)
		return baseTopic.AddTopic(topicSections[1:])
	}
}

func (topicToClient *TopicsToClientStore) GetMatchingClients(topicName string) (*structures.LinkedList[ClientID], error) {

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
	return result, nil

}

func (TopicToClient *TopicsToClientStore) get(topicName string) (*structures.LinkedList[ClientID], error) {
	topicSections := strings.Split(topicName, "/")

	topLevelTopic := TopicToClient.topLevelMap.Get(topicSections[0])
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
	fmt.Print(t.name, "(", t.connectedClients.GetItems(), ")", ": ")
	for _, child := range t.children {
		fmt.Print(child.name, " ")
	}
	fmt.Println()

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
		connectedClients: &connectedClients,
	}
	return &newTopic
}

var ErrTopicDoesntExist = errors.New("error: Topic doesn't exist")

func (t *topic) DeleteTopic(topicSections []string) error {

	if len(topicSections) == 1 {

		for i, child := range t.children {
			if child.name == topicSections[0] {
				child.children = nil
				child.connectedClients.DeleteLinkedList()
				child.connectedClients = nil

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

// TODO: Add + support
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
		if child.name == topicSections[0] || topicSections[0] == "+" {
			result = structures.Concatenate(result, child.getMatchingClients(topicSections[1:]))
		}
	}
	return result

}

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