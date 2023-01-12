package client

import (
	"MQTT-GO/structures"
	"fmt"
	"sync"
)

type packetIdStore struct {
	packetIdentifier int
	packetIdLock     sync.Mutex
}

var (
	packetIdentifier = packetIdStore{
		packetIdentifier: 0,
		packetIdLock:     sync.Mutex{},
	}
)

// func incrementPacketId() {
// 	packetIdentifier.packetIdLock.Lock()
// 	defer packetIdentifier.packetIdLock.Unlock()
// 	packetIdentifier.packetIdentifier++
// }

// func getPacketId() int {
// 	packetIdentifier.packetIdLock.Lock()
// 	defer packetIdentifier.packetIdLock.Unlock()
// 	return packetIdentifier.packetIdentifier
// }

func getAndIncrementPacketId() int {
	packetIdentifier.packetIdLock.Lock()
	defer packetIdentifier.packetIdLock.Unlock()
	packetIdentifier.packetIdentifier++
	return packetIdentifier.packetIdentifier - 1
}

type waitingPackets struct {
	packetList    *structures.LinkedList[*storedPacket]
	waitCondition *sync.Cond
}

type storedPacket struct {
	packet   []byte
	packetID int
}

func CreateWaitingPacketList() *waitingPackets {

	conditionMutex := sync.Mutex{}
	waitingPacketStruct := waitingPackets{
		waitCondition: sync.NewCond(&conditionMutex),
		packetList:    structures.CreateLinkedList[*storedPacket](),
	}

	return &waitingPacketStruct
}

func (wp *waitingPackets) AddItem(storedPacket *storedPacket) {
	wp.waitCondition.L.Lock()
	wp.packetList.Append(storedPacket)
	wp.waitCondition.Broadcast()
	wp.waitCondition.L.Unlock()
}

func (wp *waitingPackets) getItem(packetIdentifier int) *[]byte {
	packetFinder := func(s *storedPacket) bool { return s.packetID == packetIdentifier }
	packetStore := wp.packetList.FilterSingleItem(packetFinder)
	if packetStore != nil {
		return &(*packetStore).packet
	}
	return nil
}

func (wp *waitingPackets) GetOrWait(packetIdentifier int) *[]byte {

	wp.waitCondition.L.Lock()
	for {
		storedPacket := wp.getItem(packetIdentifier)
		if storedPacket == nil {
			fmt.Println()
			wp.waitCondition.Wait()
		} else {
			wp.waitCondition.L.Unlock()
			return storedPacket
		}
	}
}
