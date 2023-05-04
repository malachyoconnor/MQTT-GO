package structures

import (
	"sync"
)

// SafeMap is a thread safe map implementation.
// It is a wrapper around a map that contains a RWMutex.
type SafeMap[Key comparable, Value any] struct {
	clientTable map[Key]Value
	tableLock   sync.RWMutex
}

// Get returns the value associated with the key. If the key is not in the map, it returns nil.
// This read locks the map.
func (clientTable *SafeMap[Key, Value]) Get(key Key) Value {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()
	return clientTable.clientTable[key]
}

// Contains returns true if the key is in the map, false otherwise.
// This read locks the map.
func (clientTable *SafeMap[Key, Value]) Contains(key Key) bool {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()

	_, found := clientTable.clientTable[key]
	return found
}

// Put puts the key value pair into the map. If the key already exists, it overwrites the value.
// This write locks the map.
func (clientTable *SafeMap[Key, Value]) Put(key Key, value Value) {
	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	clientTable.clientTable[key] = value
}

// Delete deletes the key value pair from the map. If the key is not in the map, it does nothing.
// This write locks the map.
func (clientTable *SafeMap[Key, Value]) Delete(key Key) {
	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	delete(clientTable.clientTable, key)
}

// PutIfAbsent puts the key value pair into the map if the key does not already exist.
// If the key already exists, it returns the value associated with the key.
func (clientTable *SafeMap[Key, Value]) PutIfAbsent(key Key, value Value) Value {
	if clientTable.Contains(key) {
		return clientTable.Get(key)
	}
	clientTable.Put(key, value)
	return value
}

// CreateSafeMap creates a new SafeMap.
func CreateSafeMap[Key comparable, Value any]() *SafeMap[Key, Value] {
	result := SafeMap[Key, Value]{
		clientTable: make(map[Key]Value),
		tableLock:   sync.RWMutex{},
	}
	return &result
}

// Values returns a slice of all the values in the map.
func (clientTable *SafeMap[Key, Value]) Values() []Value {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()
	values := make([]Value, len(clientTable.clientTable))
	i := 0
	for _, val := range clientTable.clientTable {
		values[i] = val
		i++
	}
	return values
}
