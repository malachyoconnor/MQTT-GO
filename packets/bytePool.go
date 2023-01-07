package packets

import (
	"sync"
)

type BytePool struct {
	pool *sync.Pool
	// Simple "semaphore" representing the number of items waiting to be collected
	itemsWaiting chan struct{}
}

func CreateBytePool() *BytePool {
	return &BytePool{
		pool:         &sync.Pool{},
		itemsWaiting: make(chan struct{}, 100),
	}
}

func (b *BytePool) Put(item []byte) {
	if item == nil {
		panic("Error: Tried to put nil into the BytePool")
	}

	//lint:ignore SA6002 We want to allocate into the BytePool, that's the whole point
	b.pool.Put(item)
	b.itemsWaiting <- struct{}{}
}

func (b *BytePool) Get() []byte {
	<-b.itemsWaiting
	// Get a value and assert it's type as a byte array
	result := b.pool.Get()

	for result == nil {
		result = b.pool.Get()
	}

	return result.([]byte)
}
