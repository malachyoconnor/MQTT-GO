package gobro

import (
	"sync"
)

type BytePool struct {
	pool *sync.Pool
	// Simple semaphore representing the number of items waiting to be collected
	itemsWaiting chan struct{}
}

func CreateBytePool() *BytePool {
	return &BytePool{
		pool:         &sync.Pool{},
		itemsWaiting: make(chan struct{}, 100),
	}
}

func (b *BytePool) Put(item []byte) {
	b.pool.Put(item)
	b.itemsWaiting <- struct{}{}
}

func (b *BytePool) Get() []byte {
	<-b.itemsWaiting
	// Get a value and assert it's type as a byte array
	result := b.pool.Get().([]byte)
	return result
}
