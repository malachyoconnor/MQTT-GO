package client

import (
	"sync/atomic"
)

type clientQueue struct {
	workBeingDone atomic.Bool
	waitingList   *chan struct{}
}

func (queue *clientQueue) joinWaitList() {
	// If work is being done (if we find a true) then join the waiting queue
	// Otherwise, start doing our own work.
	if queue.workBeingDone.CompareAndSwap(false, true) {
		<-(*queue.waitingList)
	} else {
		return
	}

}
