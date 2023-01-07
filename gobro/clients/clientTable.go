package clients

import (
	"MQTT-GO/structures"
)

func CreateClientTable() *structures.SafeMap[ClientID, *Client] {
	result := *structures.CreateSafeMap[ClientID, *Client]()
	return &result
}
