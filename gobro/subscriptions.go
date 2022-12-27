package gobro

import "MQTT-GO/client"

type Topic struct {
	TopicFilter string
	Qos         byte
}
type ClientTopicMap map[client.ClientID][]Topic
type TopicClientMap map[Topic][]client.ClientID

func (CTMap *ClientTopicMap) addClientTopicPair(clientId client.ClientID, newTopic Topic) {

	for _, topic := range (*CTMap)[clientId] {
		if topic == newTopic {
			return
		}
	}

	(*CTMap)[clientId] = append((*CTMap)[clientId], newTopic)
}

func (TCMap *TopicClientMap) addTopicClientPair(topic Topic, newClientId client.ClientID) {

	// TODO: HANDLE WILDCARD TOPICS
	// We should just maintain a list of topics and find a way of querying the closest
	// one to the wildcard

	for _, clientID := range (*TCMap)[topic] {
		if newClientId == clientID {
			return
		}
	}

	(*TCMap)[topic] = append((*TCMap)[topic], newClientId)
}

func getSubscriptionsFromWildcard(topic Topic) []Topic {
	return []Topic{topic}
}
