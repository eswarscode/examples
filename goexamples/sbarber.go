package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type BarberShop struct {
	waitingRoom  chan int      // Buffered channel for waiting customers (stores customer IDs)
	barberSleep  chan struct{} // Signals when barber goes to sleep
	customerDone chan struct{} // Signals when customer's haircut is finished
	wg           sync.WaitGroup
	mutex        sync.Mutex
	isBarberBusy bool
}

func NewBarberShop(chairs int) *BarberShop {
	return &BarberShop{
		waitingRoom:  make(chan int, chairs),
		barberSleep:  make(chan struct{}, 1),
		customerDone: make(chan struct{}, 1),
		isBarberBusy: false,
	}
}

func (bs *BarberShop) barber() {
	defer bs.wg.Done()

	for {
		fmt.Println("💤 Barber is sleeping...")

		// Wait for a customer to wake up the barber
		select {
		case customerID := <-bs.waitingRoom:
			bs.mutex.Lock()
			bs.isBarberBusy = true
			bs.mutex.Unlock()

			fmt.Printf("✂️  Barber is cutting hair for Customer %d\n", customerID)

			// Simulate cutting hair (2-5 seconds)
			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

			fmt.Printf("✅ Barber finished cutting hair for Customer %d\n", customerID)

			// Signal that customer is done
			bs.customerDone <- struct{}{}

			bs.mutex.Lock()
			bs.isBarberBusy = false
			bs.mutex.Unlock()

		case <-time.After(10 * time.Second):
			fmt.Println("🏠 Barber shop is closing (no customers for too long)")
			return
		}
	}
}

func (bs *BarberShop) customer(id int) {
	defer bs.wg.Done()

	fmt.Printf("🚶 Customer %d is approaching the barber shop\n", id)

	// Try to enter the waiting room
	select {
	case bs.waitingRoom <- id:
		fmt.Printf("🪑 Customer %d is waiting (chairs occupied: %d)\n", id, len(bs.waitingRoom))

		// Wait for haircut to be completed
		<-bs.customerDone

		fmt.Printf("😊 Customer %d got haircut and left happy\n", id)

	default:
		fmt.Printf("😞 Customer %d left (waiting room is full)\n", id)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	waitingChairs := 3
	fmt.Printf("🏪 Opening the Sleeping Barber Shop with %d waiting chairs\n", waitingChairs)

	shop := NewBarberShop(waitingChairs)

	// Start the barber
	shop.wg.Add(1)
	go shop.barber()

	// Generate customers at random intervals
	customerCount := 0
	ticker := time.NewTicker(time.Millisecond * 800)
	defer ticker.Stop()

	timeout := time.After(20 * time.Second)

	for {
		select {
		case <-ticker.C:
			// Random chance of customer arriving (70% probability)
			if rand.Float32() < 0.7 {
				customerCount++
				shop.wg.Add(1)
				go shop.customer(customerCount)
			}

		case <-timeout:
			fmt.Println("\n🕐 Simulation time ended")

			// Close the waiting room to prevent new customers
			close(shop.waitingRoom)

			// Wait for all customers to finish
			shop.wg.Wait()

			fmt.Println("👋 Barber shop closed")
			return
		}
	}
}
