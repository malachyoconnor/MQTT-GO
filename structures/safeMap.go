package structures

import "sync"

type SafeMap[Key comparable, Value comparable] struct {
	clientTable map[Key]Value
	tableLock   sync.RWMutex
}

func (clientTable *SafeMap[Key, Value]) Get(key Key) Value {

	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()
	return clientTable.clientTable[key]

}

func (clientTable *SafeMap[Key, Value]) Exists(key Key) bool {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()

	_, found := clientTable.clientTable[key]
	return found
}

func (clientTable *SafeMap[Key, Value]) Put(key Key, value Value) {

	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	clientTable.clientTable[key] = value

}

func (clientTable *SafeMap[Key, Value]) Delete(key Key) {
	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	delete(clientTable.clientTable, key)
}

func CreateSafeMap[Key comparable, Value comparable]() *SafeMap[Key, Value] {
	result := SafeMap[Key, Value]{
		clientTable: make(map[Key]Value),
		tableLock:   sync.RWMutex{},
	}
	return &result
}
