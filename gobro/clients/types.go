package clients

import (
	"MQTT-GO/structures"
)

type Topic struct {
	TopicFilter string
	Qos         byte
}

type TopicToClient map[Topic]*structures.LinkedList[ClientID]

func (topicToClient *TopicToClient) Print() {
	ServerPrintf("Topic to client map: ")
	for t := range *topicToClient {
		ServerPrintf("%v : ", t)
		(*topicToClient)[t].PrintItems()
		ServerPrintln()
	}
}

func (topicToClient *TopicToClient) AddTopicClientPair(topic Topic, newClientID ClientID) {
	clientLL := (*topicToClient)[topic]
	if !clientLL.Contains(newClientID) {
		clientLL.Append(newClientID)
	}
}
