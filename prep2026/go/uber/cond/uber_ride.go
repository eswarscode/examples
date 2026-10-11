package main

import (
	"fmt"
	"sync"
)

// The Uber Ride problem, solved with ONLY a mutex and condition variables
// (no semaphores, no barrier).
//
// A car may depart only with exactly 4 riders in a balanced combination:
// 4 Democrats, 4 Republicans, or 2 Democrats + 2 Republicans. A lone minority
// (1+3 / 3+1) is never allowed. All 4 must be seated before the car departs,
// and one of them drives.
//
// Key idea: a single lock protects all state. Two condition variables let
// waiting riders of each party block until they're granted a seat ("ticket").
// Group formation is serialized by a rideBusy flag, so no illegal or
// overlapping group can form. The "all seated before departure" rule needs no
// barrier: boarding happens under the lock, so the 4th rider to board is
// trivially the last — it becomes the driver and departs.

type uberStop struct {
	mu   sync.Mutex
	cDem *sync.Cond // Democrats wait here for a seat
	cRep *sync.Cond // Republicans wait here for a seat

	waitingD int // Democrats currently waiting
	waitingR int // Republicans currently waiting

	rideBusy   bool // a group is currently boarding
	demTickets int  // seats granted to Democrats for the current group
	repTickets int  // seats granted to Republicans for the current group
	boarded    int  // how many of the current group have taken their seat

	rideID int
}

func newUberStop() *uberStop {
	u := &uberStop{}
	u.cDem = sync.NewCond(&u.mu)
	u.cRep = sync.NewCond(&u.mu)
	return u
}

// tryForm starts a new group if a legal one can be assembled. Must hold u.mu.
func (u *uberStop) tryForm() {
	if u.rideBusy {
		return
	}
	switch {
	case u.waitingD >= 4:
		u.waitingD -= 4
		u.demTickets = 4
	case u.waitingD >= 2 && u.waitingR >= 2:
		u.waitingD -= 2
		u.waitingR -= 2
		u.demTickets = 2
		u.repTickets = 2
	default:
		return // not enough riders for any legal group yet
	}
	u.rideBusy = true
	u.boarded = 0
	// Wake everyone; each rider re-checks its ticket count under the lock.
	u.cDem.Broadcast()
	u.cRep.Broadcast()
}

// board takes a seat and, if this is the 4th rider, drives. Must hold u.mu.
func (u *uberStop) board(party string, id int) {
	fmt.Printf("%s %d: seated.\n", party, id)
	u.boarded++
	if u.boarded == 4 {
		u.rideID++
		fmt.Printf(">>> Ride %d departing — driver: %s %d. <<<\n", u.rideID, party, id)
		u.rideBusy = false
		u.tryForm() // immediately try to assemble the next group
	}
}

func (u *uberStop) democrat(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.mu.Lock()
	defer u.mu.Unlock()

	u.waitingD++
	u.tryForm()
	for u.demTickets == 0 { // wait until granted a seat
		u.cDem.Wait()
	}
	u.demTickets--
	u.board("Democrat", id)
}

func (u *uberStop) republican(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.mu.Lock()
	defer u.mu.Unlock()

	u.waitingR++
	u.tryForm()
	for u.repTickets == 0 {
		u.cRep.Wait()
	}
	u.repTickets--
	u.board("Republican", id)
}

func main() {
	const dems, reps = 8, 8 // both even and >= 4 so every rider departs
	u := newUberStop()

	var wg sync.WaitGroup
	wg.Add(dems + reps)
	for i := 1; i <= dems || i <= reps; i++ {
		if i <= dems {
			go u.democrat(i, &wg)
		}
		if i <= reps {
			go u.republican(i, &wg)
		}
	}

	wg.Wait()
	fmt.Printf("All %d riders have departed in %d rides.\n", dems+reps, u.rideID)
}
