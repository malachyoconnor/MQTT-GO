package structures

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// TicketStand is a thread safe implementation of a ticket stand.
// It uses ticketnumbers to ensure that tickets are completed in order.
// It uses a sync.Cond so threads can wait on tickets, and wake up when
// the ticket is completed.
type TicketStand struct {
	earliestTicket atomic.Int64
	latestTicket   atomic.Int64
	waitingTickets *SafeMap[int64, chan struct{}]
	startTimes     *SafeMap[int64, int64]
	closed         chan struct{}
	tStandClosed   atomic.Bool
}

// CreateTicketStand creates a new TicketStand.
func CreateTicketStand() *TicketStand {

	tStand := TicketStand{
		earliestTicket: atomic.Int64{},
		latestTicket:   atomic.Int64{},
		waitingTickets: CreateSafeMap[int64, chan struct{}](),
		startTimes:     CreateSafeMap[int64, int64](),
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
	ticketNumber := tHolder.latestTicket.Add(1) - 1

	tHolder.startTimes.Put(ticketNumber, time.Now().UnixNano())

	queue := make(chan struct{}, 2)
	if ticketNumber == tHolder.earliestTicket.Load() {
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
			return
		}
	}
}

func (ticket *Ticket) StopTiming() int64 {
	startTime := ticket.ticketStand.startTimes.Get(ticket.ticketNumber)
	timeToCompleteTicket := time.Now().UnixNano() - startTime
	if len(toWriteNanos) < 10000 {
		toWriteNanos <- timeToCompleteTicket
	}
	return timeToCompleteTicket
}

var finishedWriting = make(chan struct{}, 1)
var toWriteNanos = make(chan int64, 10000)

func WriteToCsv(filename string) {

	csvFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()
	for {
		select {
		case <-finishedWriting:
			return
		case duration := <-toWriteNanos:
			{
				_, err := csvFile.Write([]byte(fmt.Sprint(duration, ",")))
				if err != nil {
					panic(err)
				}
			}
		}
	}

}

func StopWriting() {
	finishedWriting <- struct{}{}
}

// Complete completes the ticket. It increments the current ticket number
func (ticket *Ticket) Complete() {
	if ticket.ticketStand.tStandClosed.Load() {
		return
	}

	if ticket.ticketNumber != ticket.ticketStand.earliestTicket.Load() {
		fmt.Println("Tried to complete ticket at the wrong time")
		panic("??")
	}

	newTicket := ticket.ticketStand.earliestTicket.Add(1)
	waitingChannel := ticket.ticketStand.waitingTickets.Get(newTicket)
	waitingChannel <- struct{}{}

}
