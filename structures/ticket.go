package structures

import (
	"sync"
)

// TicketStand is a thread safe implementation of a ticket stand.
// It uses ticketnumbers to ensure that tickets are completed in order.
// It uses a sync.Cond so threads can wait on tickets, and wake up when
// the ticket is completed.
type TicketStand struct {
	currentTicket   int64
	nextTicket      int64
	holderLock      sync.Mutex
	ticketCompleted *sync.Cond
}

// CreateTicketStand creates a new TicketStand.
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

// GetTicket returns a new ticket. The ticket number is the next ticket number.
func (tHolder *TicketStand) GetTicket() Ticket {
	tHolder.holderLock.Lock()
	ticketNumber := tHolder.nextTicket
	tHolder.nextTicket++
	tHolder.holderLock.Unlock()

	return Ticket{
		ticketNumber: ticketNumber,
		ticketStand:  tHolder,
	}
}

// Ticket is a ticket that can be used to wait on a ticket stand.
type Ticket struct {
	ticketNumber int64
	ticketStand  *TicketStand
}

// WaitOnTicket waits on the ticket stand until the ticket is completed.
func (ticket *Ticket) WaitOnTicket() {
	ticket.ticketStand.ticketCompleted.L.Lock()
	for ticket.ticketNumber != ticket.ticketStand.currentTicket {
		ticket.ticketStand.ticketCompleted.Wait()
	}
	ticket.ticketStand.ticketCompleted.L.Unlock()
}

// TicketCompleted completes the ticket. It increments the current ticket number
func (ticket *Ticket) TicketCompleted() {
	ticket.ticketStand.ticketCompleted.L.Lock()
	ticket.ticketStand.holderLock.Lock()
	ticket.ticketStand.currentTicket++
	ticket.ticketStand.holderLock.Unlock()
	ticket.ticketStand.ticketCompleted.Broadcast()
	ticket.ticketStand.ticketCompleted.L.Unlock()
}
