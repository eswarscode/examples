package main

import (
	"fmt"
	"sync"
	"time"
)

type Barbershop struct {
	waitingRoom    chan int
	wakeUpBarber   chan struct{}
	customerDone   chan struct{}
	wg             sync.WaitGroup
	numChairs      int
	barberBusy     bool
	barberSleeping bool
	mutex          sync.Mutex
}

func NewBarbershop(numChairs int) *Barbershop {
	return &Barbershop{
		waitingRoom:    make(chan int, numChairs),
		wakeUpBarber:   make(chan struct{}, 1),
		customerDone:   make(chan struct{}),
		numChairs:      numChairs,
		barberBusy:     false,
		barberSleeping: false,
	}
}

func (bs *Barbershop) customer(customerID int) {
	defer bs.wg.Done()

	fmt.Printf("Customer %d: Entering barbershop\n", customerID)

	select {
	case bs.waitingRoom <- customerID:
		fmt.Printf("Customer %d: Got a seat, waiting...\n", customerID)
		time.Sleep(100 * time.Millisecond) // Simulate some waiting
	default:
		fmt.Printf("Customer %d: All chairs occupied, leaving without haircut\n", customerID)
	}
}

func main() {
	numChairs := 2
	barbershop := NewBarbershop(numChairs)

	fmt.Printf("Test: Barbershop with %d chairs\n", numChairs)

	// Send customers rapidly to fill up waiting room
	for i := 1; i <= 5; i++ {
		barbershop.wg.Add(1)
		go barbershop.customer(i)
		time.Sleep(10 * time.Millisecond) // Very short delay
	}

	barbershop.wg.Wait()
	fmt.Println("Test completed")
}
