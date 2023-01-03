package clients

import "sync"

type ClientTable struct {
	clientTable map[ClientID]*Client
	tableLock   sync.RWMutex
}

func (clientTable *ClientTable) Get(key ClientID) *Client {

	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()
	return clientTable.clientTable[key]

}

func (clientTable *ClientTable) Exists(key ClientID) bool {
	clientTable.tableLock.RLock()
	defer clientTable.tableLock.RUnlock()

	_, found := clientTable.clientTable[key]
	return found
}

func (clientTable *ClientTable) Put(key ClientID, value *Client) {

	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	clientTable.clientTable[key] = value

}

func (clientTable *ClientTable) Delete(key ClientID) {
	clientTable.tableLock.Lock()
	defer clientTable.tableLock.Unlock()
	delete(clientTable.clientTable, key)
}

func CreateClientTable() *ClientTable {

	result := ClientTable{
		clientTable: make(map[ClientID]*Client),
		tableLock:   sync.RWMutex{},
	}
	return &result

}
