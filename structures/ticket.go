package structures

import (
	"sync"
)

type TicketStand struct {
	currentTicket   int64
	nextTicket      int64
	holderLock      sync.Mutex
	ticketCompleted *sync.Cond
}

func CreateTicketStand() *TicketStand {
	conditionMutex := sync.Mutex{}

	tStand := TicketStand{
		currentTicket:   0,
		nextTicket:      0,
		holderLock:      sync.Mutex{},
		ticketCompleted: sync.NewCond(&conditionMutex),
	}

	return &tStand
}

func (tHolder *TicketStand) GetTicket() Ticket {
	tHolder.holderLock.Lock()
	ticketNumber := tHolder.nextTicket
	tHolder.nextTicket += 1
	tHolder.holderLock.Unlock()

	return Ticket{
		ticketNumber: ticketNumber,
		ticketStand:  tHolder,
	}
}

type Ticket struct {
	ticketNumber int64
	ticketStand  *TicketStand
}

func (ticket *Ticket) WaitOnTicket() {
	ticket.ticketStand.ticketCompleted.L.Lock()
	for ticket.ticketNumber != ticket.ticketStand.currentTicket {
		ticket.ticketStand.ticketCompleted.Wait()
	}
	ticket.ticketStand.ticketCompleted.L.Unlock()
}

func (ticket *Ticket) TicketCompleted() {
	ticket.ticketStand.ticketCompleted.L.Lock()
	ticket.ticketStand.holderLock.Lock()
	ticket.ticketStand.currentTicket += 1
	ticket.ticketStand.holderLock.Unlock()
	ticket.ticketStand.ticketCompleted.Broadcast()
	ticket.ticketStand.ticketCompleted.L.Unlock()
}
