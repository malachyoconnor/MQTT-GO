package clients

type Topic struct {
	TopicFilter string
	Qos         byte
}

type ClientIDNode struct {
	clientID ClientID
	nextNode *ClientIDNode
	prevNode *ClientIDNode
}

type TopicClientMap map[Topic]ClientIDNode

func (TCMap *TopicClientMap) AddTopicClientPair(topic Topic, newClientId ClientID) {

	// TODO: HANDLE WILDCARD TOPICS
	// We should just maintain a list of topics and find a way of querying the closest
	// one to the wildcard

	// for

	// for _, clientID := range (*TCMap)[topic] {
	// 	if newClientId == clientID {
	// 		return
	// 	}
	// }

	// (*TCMap)[topic] = append((*TCMap)[topic], newClientId)
}

func getSubscriptionsFromWildcard(topic Topic) []Topic {
	return []Topic{topic}
}
