package structures

import "sync"

type SafeMap[Key comparable, Value any] struct {
	clientTable map[Key]Value
	tableLock   sync.RWMutex
}

func (clientTable *SafeMap[Key, Value]) Get(key Key) Value {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()
	return clientTable.clientTable[key]
}

func (clientTable *SafeMap[Key, Value]) Contains(key Key) bool {
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

func (clientTable *SafeMap[Key, Value]) PutIfAbsent(key Key, value Value) Value {
	if clientTable.Contains(key) {
		return clientTable.Get(key)
	} else {
		clientTable.Put(key, value)
		return value
	}
}

func CreateSafeMap[Key comparable, Value any]() *SafeMap[Key, Value] {
	result := SafeMap[Key, Value]{
		clientTable: make(map[Key]Value),
		tableLock:   sync.RWMutex{},
	}
	return &result
}

func (table *SafeMap[Key, Value]) Values() []Value {
	values := make([]Value, len(table.clientTable))
	i := 0
	for _, val := range table.clientTable {
		values[i] = val
		i++
	}
	return values
}
