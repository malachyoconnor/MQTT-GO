package gobro

import "MQTT-GO/client"

type Topic string
type SubscriptionTable map[client.ClientID][]Topic
