package clients

import (
	"sync/atomic"
)

type ClientQueue struct {
	WorkBeingDone atomic.Bool
	WaitingList   *chan struct{}
}

func (queue *ClientQueue) JoinWaitList() {
	// If work is being done (if we find a true) then join the waiting queue
	// Otherwise, start doing our own work.
	if len(*queue.WaitingList) > 0 {
		<-(*queue.WaitingList)
	}

	for !queue.WorkBeingDone.CompareAndSwap(false, true) {
		<-(*queue.WaitingList)
	}
}

func (queue *ClientQueue) DoingWork() {
	// If work is being done but we don't need to wait for the preceeding work to finish
	queue.WorkBeingDone.CompareAndSwap(false, true)
}

func (queue *ClientQueue) FinishedWork() {
	if queue.WorkBeingDone.CompareAndSwap(true, false) {
		<-(*queue.WaitingList)
	} else {
		panic("Finished work on queue without work being done!")
	}
}
