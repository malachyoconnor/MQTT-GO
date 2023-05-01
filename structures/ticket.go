package structures

import (
	"fmt"
	"sync/atomic"
)

// TicketStand is a thread safe implementation of a ticket stand.
// It uses ticketnumbers to ensure that tickets are completed in order.
// It uses a sync.Cond so threads can wait on tickets, and wake up when
// the ticket is completed.
type TicketStand struct {
	currentTicket  atomic.Int64
	nextTicket     atomic.Int64
	waitingTickets *SafeMap[int64, chan struct{}]
	closed         chan struct{}
	tStandClosed   atomic.Bool
}

// CreateTicketStand creates a new TicketStand.
func CreateTicketStand() *TicketStand {

	tStand := TicketStand{
		currentTicket:  atomic.Int64{},
		nextTicket:     atomic.Int64{},
		waitingTickets: CreateSafeMap[int64, chan struct{}](),
		tStandClosed:   atomic.Bool{},
		closed:         make(chan struct{}, 1),
	}

	return &tStand
}

func (tHolder *TicketStand) CloseTicketStand() error {
	if tHolder.tStandClosed.Load() {
		return nil
	}
	tHolder.tStandClosed.Store(true)
	close(tHolder.closed)

	defer func() {
		recover()
	}()

	queueList := tHolder.waitingTickets.Values()
	for _, queue := range queueList {
		close(queue)
	}
	return nil
}

// GetTicket returns a new ticket. The ticket number is the next ticket number.
func (tHolder *TicketStand) GetTicket() Ticket {
	ticketNumber := tHolder.nextTicket.Add(1) - 1

	queue := make(chan struct{}, 1)
	if ticketNumber == tHolder.currentTicket.Load() {
		queue <- struct{}{}
	}

	tHolder.waitingTickets.Put(ticketNumber, queue)
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

// Wait waits on the ticket stand until the ticket can be
// safely completed
func (ticket *Ticket) Wait() {
	if ticket.ticketStand.tStandClosed.Load() {
		return
	}
	select {
	case <-ticket.ticketStand.closed:
		{
			return
		}
	// Wait until the previous ticket completes and
	// sends a signal on the channel
	case <-ticket.ticketStand.waitingTickets.Get(ticket.ticketNumber):
		{
			break
		}
	}
}

// Complete completes the ticket. It increments the current ticket number
func (ticket *Ticket) Complete() {
	if ticket.ticketStand.tStandClosed.Load() {
		return
	}
	defer func() {
		recover()
	}()
	if ticket.ticketNumber != ticket.ticketStand.currentTicket.Load() {
		fmt.Println("Tried to complete ticket at the wrong time")
		return
	}
	newTicket := ticket.ticketStand.currentTicket.Add(1)
	waitingChannel := ticket.ticketStand.waitingTickets.Get(newTicket)
	waitingChannel <- struct{}{}

}
