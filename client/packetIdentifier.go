package client

import (
	"sync"
	"sync/atomic"

	"MQTT-GO/structures"
)

type packetIDStore struct {
	packetIdentifier atomic.Int64
}

var packetIdentifier = packetIDStore{
	packetIdentifier: atomic.Int64{},
}

func getAndIncrementPacketID() int {
	return int(packetIdentifier.packetIdentifier.Add(1))
}

// WaitingAcks is a struct that stores a list of packets that are waiting for an ACK.
// It uses a sync.Cond to wait for the ACK. And broadcasts when an ACK is added.
// Waiting threads then wake up and check if their packet has been added
type WaitingAcks struct {
	PacketList    *structures.LinkedList[*StoredPacket]
	waitCondition *sync.Cond
}

// StoredPacket is a struct that stores a packet, and the packet identifier.
// This is used to store packets that are waiting for an ACK.
type StoredPacket struct {
	Packet   []byte
	PacketID int
}

// CreateWaitingAckList creates a new waitingAcks struct
func CreateWaitingAckList() *WaitingAcks {
	conditionMutex := sync.Mutex{}
	waitingPacketStruct := WaitingAcks{
		waitCondition: sync.NewCond(&conditionMutex),
		PacketList:    structures.CreateLinkedList[*StoredPacket](),
	}
	return &waitingPacketStruct
}

// AddItem adds a packet to the list of packets that are waiting for an ACK.
func (wp *WaitingAcks) AddItem(storedPacket *StoredPacket) {
	wp.waitCondition.L.Lock()
	wp.PacketList.Append(storedPacket)
	wp.waitCondition.Broadcast()
	wp.waitCondition.L.Unlock()
}

func (wp *WaitingAcks) getItem(packetIdentifier int) *[]byte {
	packetFinder := func(s *StoredPacket) bool { return s.PacketID == packetIdentifier }
	packetStore := wp.PacketList.FilterSingleItem(packetFinder)
	if packetStore != nil {
		return &(*packetStore).Packet
	}
	return nil
}

// GetOrWait gets a packet from the list of packets that are waiting for an ACK.
// If the packet is not in the list, it waits for a broadcast from the AddItem function.
func (wp *WaitingAcks) GetOrWait(packetIdentifier int) *[]byte {
	wp.waitCondition.L.Lock()
	defer wp.waitCondition.L.Unlock()
	for {
		storedPacket := wp.getItem(packetIdentifier)
		if storedPacket == nil {
			structures.Println()
			wp.waitCondition.Wait()
		} else {
			return storedPacket
		}
	}
}
