package clients

import (
	"MQTT-GO/structures"
)

type Topic struct {
	TopicFilter string
	Qos         byte
}

type ClientIDNode struct {
	clientID ClientID
	nextNode *ClientIDNode
	prevNode *ClientIDNode
}

type TopicToClient map[Topic]structures.LinkedList[ClientID]

func (topicToClient *TopicToClient) AddTopicClientPair(topic Topic, newClientId ClientID) {

	// TODO: HANDLE WILDCARD TOPICS
	// We should just maintain a list of topics and find a way of querying the closest
	// one to the wildcard

	clientLL := (*topicToClient)[topic]
	if clientLL.Contains(newClientId) {
		return
	}
	clientLL.Append(newClientId)
}

func getSubscriptionsFromWildcard(topic Topic) []Topic {
	return []Topic{topic}
}
