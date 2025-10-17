package main

import (
	"fmt"
	"sync"
	"time"
)

const maxCapacity = 3

// Bathroom is the shared resource with synchronization mechanisms.
type Bathroom struct {
	mu          sync.Mutex
	cond        *sync.Cond
	occupants   int
	genderInUse string // "male", "female", or ""
}

// NewBathroom creates and returns a new Bathroom instance.
func NewBathroom() *Bathroom {
	b := &Bathroom{}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Enter attempts to enter the bathroom.
func (b *Bathroom) Enter(gender string, id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Wait until it's safe to enter.
	for b.occupants == maxCapacity || (b.genderInUse != "" && b.genderInUse != gender) {
		b.cond.Wait()
	}

	// Conditions are met; enter the bathroom.
	b.occupants++
	b.genderInUse = gender
	fmt.Printf("%s %d has entered. Occupants: %d\n", gender, id, b.occupants)
}

// Exit leaves the bathroom.
func (b *Bathroom) Exit(gender string, id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.occupants--
	fmt.Printf("%s %d has exited. Occupants: %d\n", gender, id, b.occupants)

	if b.occupants == 0 {
		b.genderInUse = "" // Bathroom is empty; reset gender.
	}

	// Wake up everyone. The condition loop will make them re-check.
	b.cond.Broadcast()
}

// Simulate an employee using the bathroom.
func useBathroom(b *Bathroom, gender string, id int) {
	b.Enter(gender, id)
	time.Sleep(time.Second) // Simulate using the bathroom.
	b.Exit(gender, id)
}

func main() {
	b := NewBathroom()
	var wg sync.WaitGroup

	// Start some male and female goroutines.
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			useBathroom(b, "male", id)
		}(i)

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			useBathroom(b, "female", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("All employees have finished using the bathroom.")
}
