package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// The Uber Ride problem.
//
// At the end of a political conference, Democrats and Republicans are leaving
// in Ubers. To keep the peace, a car may only depart when it holds exactly 4
// riders in one of these balanced combinations:
//   - 4 Democrats, or
//   - 4 Republicans, or
//   - 2 Democrats + 2 Republicans.
//
// A lone minority (1+3 or 3+1) is never allowed. When a valid group forms, the
// rider who completes it becomes the "captain" and drives; all 4 must be seated
// before the car departs.
//
// Classic solution: a mutex guards the waiting counts; two "queue" semaphores
// block riders until a valid group can form; and a reusable barrier makes all 4
// riders rendezvous before the captain drives. The captain holds the mutex from
// the moment the group forms until the car departs, so no second group can form
// concurrently.

// semaphore is a counting semaphore backed by a buffered channel.
type semaphore chan struct{}

func newSemaphore(capacity int) semaphore { return make(semaphore, capacity) }
func (s semaphore) acquire()              { <-s }
func (s semaphore) release(n int) {
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
}

// barrier is a reusable barrier for exactly n goroutines.
type barrier struct {
	mu    sync.Mutex
	cond  *sync.Cond
	n     int
	count int
	phase uint64
}

func newBarrier(n int) *barrier {
	b := &barrier{n: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *barrier) await() {
	b.mu.Lock()
	defer b.mu.Unlock()
	p := b.phase
	b.count++
	if b.count == b.n {
		b.count = 0
		b.phase++
		b.cond.Broadcast()
		return
	}
	for p == b.phase {
		b.cond.Wait()
	}
}

type uberStop struct {
	mu          sync.Mutex // the "mutex" semaphore guarding the counts below
	democrats   int        // Democrats currently waiting
	republicans int        // Republicans currently waiting

	demsQueue semaphore // Democrats block here until released into a group
	repsQueue semaphore // Republicans block here until released into a group
	barrier   *barrier  // all 4 group members rendezvous here before departure

	rideID int64 // monotonic ride counter (atomic)
}

func newUberStop(totalRiders int) *uberStop {
	return &uberStop{
		demsQueue: newSemaphore(totalRiders),
		repsQueue: newSemaphore(totalRiders),
		barrier:   newBarrier(4),
	}
}

func (u *uberStop) seated(party string, id int) {
	fmt.Printf("%s %d: seated.\n", party, id)
}

func (u *uberStop) drive(party string, id int) {
	ride := atomic.AddInt64(&u.rideID, 1)
	fmt.Printf(">>> Ride %d departing — captain: %s %d. <<<\n", ride, party, id)
}

func (u *uberStop) democrat(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.mu.Lock()
	captain := false
	u.democrats++
	if u.democrats == 4 {
		// Four Democrats present: release all four (3 waiting + me).
		u.demsQueue.release(4)
		u.democrats -= 4
		captain = true
	} else if u.democrats == 2 && u.republicans >= 2 {
		// Two of each present: release 2 Democrats and 2 Republicans.
		u.demsQueue.release(2)
		u.democrats -= 2
		u.repsQueue.release(2)
		u.republicans -= 2
		captain = true
	} else {
		// Can't form a group yet: release the mutex and wait.
		u.mu.Unlock()
	}

	u.demsQueue.acquire() // captain's own permit, or wait to be released
	u.seated("Democrat", id)
	u.barrier.await()
	if captain {
		u.drive("Democrat", id)
		u.mu.Unlock() // held since the group formed
	}
}

func (u *uberStop) republican(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	u.mu.Lock()
	captain := false
	u.republicans++
	if u.republicans == 4 {
		u.repsQueue.release(4)
		u.republicans -= 4
		captain = true
	} else if u.republicans == 2 && u.democrats >= 2 {
		u.repsQueue.release(2)
		u.republicans -= 2
		u.demsQueue.release(2)
		u.democrats -= 2
		captain = true
	} else {
		u.mu.Unlock()
	}

	u.repsQueue.acquire()
	u.seated("Republican", id)
	u.barrier.await()
	if captain {
		u.drive("Republican", id)
		u.mu.Unlock()
	}
}

func main() {
	const dems, reps = 8, 8 // both even and >= 4 so every rider eventually departs
	u := newUberStop(dems + reps)

	var wg sync.WaitGroup
	wg.Add(dems + reps)

	// Interleave arrivals so groups form from a mix of parties.
	for i := 1; i <= dems || i <= reps; i++ {
		if i <= dems {
			go u.democrat(i, &wg)
		}
		if i <= reps {
			go u.republican(i, &wg)
		}
	}

	wg.Wait()
	fmt.Printf("All %d riders have departed in %d rides.\n", dems+reps, atomic.LoadInt64(&u.rideID))
}
