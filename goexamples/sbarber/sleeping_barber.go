package main

import (
	"fmt"
	"math/rand"
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

func (bs *Barbershop) barber() {
	defer bs.wg.Done()

mainLoop:
	for {
		// Barber goes to sleep when no customers
		bs.mutex.Lock()
		bs.barberSleeping = true
		bs.barberBusy = false
		bs.mutex.Unlock()

		fmt.Println("💤 Barber: Going to sleep...")

		// Barber actually sleeps until woken up
		select {
		case <-bs.wakeUpBarber:
			fmt.Println("👋 Barber: Woken up!")

			// Now process all customers from waiting room
			for {
				select {
				case customerID := <-bs.waitingRoom:
					bs.mutex.Lock()
					bs.barberBusy = true
					bs.mutex.Unlock()

					fmt.Printf("✂️  Barber: Cutting hair for customer %d\n", customerID)
					time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
					fmt.Printf("✅ Barber: Finished cutting hair for customer %d\n", customerID)

					bs.customerDone <- struct{}{}

					bs.mutex.Lock()
					bs.barberBusy = false
					bs.mutex.Unlock()

				default:
					// No more customers, go back to sleep
					continue mainLoop
				}
			}

		case <-time.After(10 * time.Second):
			fmt.Println("🏠 Barber: No customers for too long, closing shop")
			return
		}
	}
}

func (bs *Barbershop) customer(customerID int) {
	defer bs.wg.Done()

	fmt.Printf("Customer %d: Entering barbershop\n", customerID)

	select {
	case bs.waitingRoom <- customerID:
		bs.mutex.Lock()
		isSleeping := bs.barberSleeping
		isBusy := bs.barberBusy

		// Atomically check and wake barber if sleeping
		if isSleeping {
			fmt.Printf("Customer %d: Found barber sleeping, waking up the barber\n", customerID)
			bs.barberSleeping = false // Mark as not sleeping immediately
			bs.mutex.Unlock()

			// Wake up the barber (non-blocking since channel is buffered)
			bs.wakeUpBarber <- struct{}{}
		} else {
			bs.mutex.Unlock()

			if isBusy {
				fmt.Printf("Customer %d: Barber is busy, waiting in chair (%d/%d chairs occupied)\n",
					customerID, len(bs.waitingRoom), bs.numChairs)
			} else {
				fmt.Printf("Customer %d: Waiting in chair (%d/%d chairs occupied)\n",
					customerID, len(bs.waitingRoom), bs.numChairs)
			}
		}

		<-bs.customerDone
		fmt.Printf("Customer %d: Leaving with fresh haircut\n", customerID)

	default:
		fmt.Printf("Customer %d: All chairs occupied, leaving without haircut\n", customerID)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numChairs := 3
	numCustomers := 8

	barbershop := NewBarbershop(numChairs)

	fmt.Printf("Barbershop opened with %d waiting chairs\n", numChairs)

	barbershop.wg.Add(1)
	go barbershop.barber()

	for i := 1; i <= numCustomers; i++ {
		barbershop.wg.Add(1)
		go barbershop.customer(i)

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
	}

	barbershop.wg.Wait()
	fmt.Println("Barbershop closed")
}
