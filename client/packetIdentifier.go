package client

import (
	"fmt"
	"sync"

	"MQTT-GO/structures"
)

type packetIdStore struct {
	packetIdentifier int
	packetIdLock     sync.Mutex
}

var packetIdentifier = packetIdStore{
	packetIdentifier: 0,
	packetIdLock:     sync.Mutex{},
}

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
	packetList    *structures.LinkedList[*StoredPacket]
	waitCondition *sync.Cond
}

type StoredPacket struct {
	Packet   []byte
	PacketID int
}

func CreateWaitingPacketList() *waitingPackets {
	conditionMutex := sync.Mutex{}
	waitingPacketStruct := waitingPackets{
		waitCondition: sync.NewCond(&conditionMutex),
		packetList:    structures.CreateLinkedList[*StoredPacket](),
	}

	return &waitingPacketStruct
}

func (wp *waitingPackets) AddItem(storedPacket *StoredPacket) {
	wp.waitCondition.L.Lock()
	wp.packetList.Append(storedPacket)
	wp.waitCondition.Broadcast()
	wp.waitCondition.L.Unlock()
}

func (wp *waitingPackets) getItem(packetIdentifier int) *[]byte {
	packetFinder := func(s *StoredPacket) bool { return s.PacketID == packetIdentifier }
	packetStore := wp.packetList.FilterSingleItem(packetFinder)
	if packetStore != nil {
		return &(*packetStore).Packet
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
