package main

import (
	"fmt"
	"sync"
	"time"
)

// The Sleeping Barber problem implemented with a mutex + condition variables
// (sync.Mutex and sync.Cond) rather than channels.
//
// A barber shop has one barber, one barber chair, and a waiting room with a
// fixed number of chairs.
//   - If there are no customers, the barber sleeps (Cond.Wait).
//   - A customer wakes the barber by signaling the condition.
//   - If the barber is busy and a waiting chair is free, the customer waits.
//   - If all waiting chairs are full, the customer leaves (balks).
//
// All shared state (waitingCount, the per-customer handoff flags, and the
// open flag) is protected by a single mutex. Three condition variables layered
// on that mutex replace the three semaphores a classic solution would use.

type shop struct {
	mu            sync.Mutex
	waitingChairs int
	queue         []*customer // FIFO of seated, waiting customers
	open          bool

	// Condition variables, all backed by mu.
	customerArrived *sync.Cond // barber waits here when the shop is empty
}

func newShop(waitingChairs int) *shop {
	s := &shop{
		waitingChairs: waitingChairs,
		open:          true,
	}
	s.customerArrived = sync.NewCond(&s.mu)
	return s
}

// customer carries its own handoff state so the barber can hand back a
// specific person rather than "whoever".
type customer struct {
	id int

	// Guarded by shop.mu via the condition variables below.
	inChair     bool       // barber has invited this customer into the chair
	cutDone     bool       // barber has finished this customer's haircut
	seatReady   *sync.Cond // customer waits here until inChair becomes true
	cutFinished *sync.Cond // customer waits here until cutDone becomes true
}

func (s *shop) newCustomer(id int) *customer {
	c := &customer{id: id}
	c.seatReady = sync.NewCond(&s.mu)
	c.cutFinished = sync.NewCond(&s.mu)
	return c
}

// barber loops: sleep until a customer arrives, serve them, repeat. Exits once
// the shop is closed and the waiting room is empty.
func (s *shop) barber(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		s.mu.Lock()

		// Sleep while there is nobody waiting and the shop is still open.
		for len(s.queue) == 0 && s.open {
			s.customerArrived.Wait()
		}

		// Closed and nobody left to serve -> go home.
		if len(s.queue) == 0 && !s.open {
			s.mu.Unlock()
			break
		}

		// Take the next waiting customer out of the waiting room (FIFO).
		c := s.queue[0]
		s.queue = s.queue[1:]

		// Invite them into the chair and wake them.
		c.inChair = true
		c.seatReady.Signal()
		s.mu.Unlock()

		// Cut hair outside the lock so other customers can still sit down.
		fmt.Printf("Barber: cutting hair for customer %d...\n", c.id)
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("Barber: finished customer %d.\n", c.id)

		// Signal the customer that the cut is complete.
		s.mu.Lock()
		c.cutDone = true
		c.cutFinished.Signal()
		s.mu.Unlock()
	}
	fmt.Println("Barber: no more customers, going home.")
}

// visit models a customer arriving. Returns true if served, false if balked.
func (s *shop) visit(c *customer) bool {
	s.mu.Lock()
	if len(s.queue) >= s.waitingChairs || !s.open {
		s.mu.Unlock()
		fmt.Printf("Customer %d: shop full, leaving.\n", c.id)
		return false
	}

	// Sit down in the waiting room and wake the barber if he's asleep.
	s.queue = append(s.queue, c)
	fmt.Printf("Customer %d: sitting down to wait (waiting=%d).\n", c.id, len(s.queue))
	s.customerArrived.Signal()

	// Wait until the barber invites us into the chair.
	for !c.inChair {
		c.seatReady.Wait()
	}
	fmt.Printf("Customer %d: getting a haircut.\n", c.id)

	// Wait until the haircut is finished.
	for !c.cutDone {
		c.cutFinished.Wait()
	}
	s.mu.Unlock()

	fmt.Printf("Customer %d: leaving, haircut done.\n", c.id)
	return true
}

// close shuts the shop: no new customers are seated, and the sleeping barber
// is woken so he can notice the shop is closed and exit.
func (s *shop) close() {
	s.mu.Lock()
	s.open = false
	s.customerArrived.Broadcast()
	s.mu.Unlock()
}

func main() {
	s := newShop(3) // 3 waiting chairs

	var barberWg sync.WaitGroup
	barberWg.Add(1)
	go s.barber(&barberWg)

	// Customers arrive faster than the barber works, so some will balk.
	var customersWg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		c := s.newCustomer(i)
		customersWg.Add(1)
		go func(c *customer) {
			defer customersWg.Done()
			s.visit(c)
		}(c)
		time.Sleep(40 * time.Millisecond)
	}

	// Let in-flight customers finish, close up, wait for the barber.
	customersWg.Wait()
	s.close()
	barberWg.Wait()
	fmt.Println("Shop closed.")
}
