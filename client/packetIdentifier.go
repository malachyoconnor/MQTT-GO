package client

import (
	"fmt"
	"sync"

	"MQTT-GO/structures"
)

type packetIDStore struct {
	packetIdentifier int
	packetIDLock     sync.Mutex
}

var packetIdentifier = packetIDStore{
	packetIdentifier: 0,
	packetIDLock:     sync.Mutex{},
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

func getAndIncrementPacketID() int {
	packetIdentifier.packetIDLock.Lock()
	defer packetIdentifier.packetIDLock.Unlock()
	packetIdentifier.packetIdentifier++
	return packetIdentifier.packetIdentifier - 1
}

type WaitingPackets struct {
	PacketList    *structures.LinkedList[*StoredPacket]
	waitCondition *sync.Cond
}

type StoredPacket struct {
	Packet   []byte
	PacketID int
}

func CreateWaitingPacketList() *WaitingPackets {
	conditionMutex := sync.Mutex{}
	waitingPacketStruct := WaitingPackets{
		waitCondition: sync.NewCond(&conditionMutex),
		PacketList:    structures.CreateLinkedList[*StoredPacket](),
	}

	return &waitingPacketStruct
}

func (wp *WaitingPackets) AddItem(storedPacket *StoredPacket) {
	wp.waitCondition.L.Lock()
	wp.PacketList.Append(storedPacket)
	wp.waitCondition.Broadcast()
	wp.waitCondition.L.Unlock()
}

func (wp *WaitingPackets) getItem(packetIdentifier int) *[]byte {
	packetFinder := func(s *StoredPacket) bool { return s.PacketID == packetIdentifier }
	packetStore := wp.PacketList.FilterSingleItem(packetFinder)
	if packetStore != nil {
		return &(*packetStore).Packet
	}
	return nil
}

func (wp *WaitingPackets) GetOrWait(packetIdentifier int) *[]byte {
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
