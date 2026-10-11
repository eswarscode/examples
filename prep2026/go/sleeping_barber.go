package main

import (
	"fmt"
	"sync"
	"time"
)

// The Sleeping Barber problem, done the Go way: channels instead of semaphores.
//
// A barber shop has one barber, one barber chair, and a waiting room with a
// fixed number of chairs.
//   - If there are no customers, the barber sleeps (blocks on the channel).
//   - A customer wakes the barber by sending into the channel.
//   - If the barber is busy and a waiting chair is free, the customer waits.
//   - If all waiting chairs are full, the customer leaves (balks).
//
// The buffered channel IS the waiting room: its capacity is the number of
// waiting chairs, and a non-blocking send models "sit down if there's room,
// otherwise leave."

type shop struct {
	// waitingRoom is the buffered channel of customers. Its buffer size is the
	// number of waiting chairs. A customer in the buffer is "seated and waiting".
	waitingRoom chan *customer
}

type customer struct {
	id int
	// done is closed by the barber when this customer's haircut is finished,
	// turning the handoff into a true rendezvous.
	done chan struct{}
}

func newShop(waitingChairs int) *shop {
	return &shop{
		waitingRoom: make(chan *customer, waitingChairs),
	}
}

// barber runs until the waiting room channel is closed and drained. It sleeps
// (blocks on the range) whenever no customer is seated.
func (s *shop) barber(wg *sync.WaitGroup) {
	defer wg.Done()
	for c := range s.waitingRoom {
		fmt.Printf("Barber: cutting hair for customer %d...\n", c.id)
		time.Sleep(100 * time.Millisecond) // simulate the haircut
		fmt.Printf("Barber: finished customer %d.\n", c.id)
		close(c.done) // tell the customer the cut is done
	}
	fmt.Println("Barber: no more customers, going home.")
}

// visit models a customer arriving. Returns true if they got (or will get) a
// haircut, false if they balked because the shop was full.
func (s *shop) visit(c *customer) bool {
	select {
	case s.waitingRoom <- c:
		// Took a waiting chair (or woke the sleeping barber). Now wait for the
		// barber to finish our cut.
		fmt.Printf("Customer %d: sitting down to wait.\n", c.id)
		<-c.done
		fmt.Printf("Customer %d: leaving, haircut done.\n", c.id)
		return true
	default:
		// Buffer full: no free chair, leave immediately.
		fmt.Printf("Customer %d: shop full, leaving.\n", c.id)
		return false
	}
}

// close shuts the shop: no new customers can be seated and the barber exits
// once the waiting room is drained.
func (s *shop) close() {
	close(s.waitingRoom)
}

func main() {
	s := newShop(3) // 3 waiting chairs

	var barberWg sync.WaitGroup
	barberWg.Add(1)
	go s.barber(&barberWg)

	// Customers arrive faster than the barber works, so some will balk.
	var customersWg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		c := &customer{id: i, done: make(chan struct{})}
		customersWg.Add(1)
		go func(c *customer) {
			defer customersWg.Done()
			s.visit(c)
		}(c)
		time.Sleep(40 * time.Millisecond)
	}

	// Let in-flight customers finish, then close up and wait for the barber.
	customersWg.Wait()
	s.close()
	barberWg.Wait()
	fmt.Println("Shop closed.")
}
