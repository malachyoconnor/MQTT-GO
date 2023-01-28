package clients

import (
	"fmt"

	"MQTT-GO/structures"
)

type Topic struct {
	TopicFilter string
	Qos         byte
}

type TopicToClient map[Topic]*structures.LinkedList[ClientID]

func (topicToClient *TopicToClient) Print() {
	fmt.Print("Topic to client map: ")
	for t := range *topicToClient {
		fmt.Print(t, ": ")
		(*topicToClient)[t].PrintItems()
		fmt.Println()
	}
}

func (topicToClient *TopicToClient) AddTopicClientPair(topic Topic, newClientID ClientID) {
	clientLL := (*topicToClient)[topic]
	if !clientLL.Contains(newClientID) {
		clientLL.Append(newClientID)
	}
}
