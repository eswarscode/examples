package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SleepingBarber struct {
	waitingChairs  int
	waitingRoom    []int // Slice to hold customer IDs
	mutex          sync.Mutex
	barberSleeping bool
	barberBusy     bool
	customerWait   *sync.Cond // Condition variable for customers to wait
	barberWake     *sync.Cond // Condition variable to wake barber
	isOpen         bool
	wg             sync.WaitGroup
}

func NewSleepingBarber(chairs int) *SleepingBarber {
	sb := &SleepingBarber{
		waitingChairs:  chairs,
		waitingRoom:    make([]int, 0, chairs),
		barberSleeping: true,
		barberBusy:     false,
		isOpen:         true,
	}
	sb.customerWait = sync.NewCond(&sb.mutex)
	sb.barberWake = sync.NewCond(&sb.mutex)
	return sb
}

func (sb *SleepingBarber) barberWork() {
	defer sb.wg.Done()

	for {
		sb.mutex.Lock()

		// Wait for customers or shop closure
		for len(sb.waitingRoom) == 0 && sb.isOpen {
			fmt.Println("💤 Barber is sleeping...")
			sb.barberSleeping = true
			sb.barberWake.Wait() // Wait for customer to wake barber
		}

		// Check if shop is closed and no customers waiting
		if !sb.isOpen && len(sb.waitingRoom) == 0 {
			sb.mutex.Unlock()
			fmt.Println("🏠 Barber is going home")
			return
		}

		// Take customer from waiting room
		if len(sb.waitingRoom) > 0 {
			customerID := sb.waitingRoom[0]
			sb.waitingRoom = sb.waitingRoom[1:]
			sb.barberSleeping = false
			sb.barberBusy = true

			fmt.Printf("✂️  Barber is cutting hair for Customer %d (waiting: %d)\n",
				customerID, len(sb.waitingRoom))

			sb.mutex.Unlock()

			// Simulate cutting hair (2-5 seconds)
			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

			sb.mutex.Lock()
			sb.barberBusy = false
			fmt.Printf("✅ Barber finished cutting hair for Customer %d\n", customerID)

			// Signal that haircut is done
			sb.customerWait.Broadcast()
			sb.mutex.Unlock()
		} else {
			sb.mutex.Unlock()
		}
	}
}

func (sb *SleepingBarber) customerArrival(customerID int) {
	defer sb.wg.Done()

	fmt.Printf("🚶 Customer %d is approaching the barber shop\n", customerID)

	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	// Check if shop is closed
	if !sb.isOpen {
		fmt.Printf("🚪 Customer %d found the shop closed\n", customerID)
		return
	}

	// Check if waiting room is full
	if len(sb.waitingRoom) >= sb.waitingChairs {
		fmt.Printf("😞 Customer %d left (waiting room is full)\n", customerID)
		return
	}

	// Add customer to waiting room
	sb.waitingRoom = append(sb.waitingRoom, customerID)
	fmt.Printf("🪑 Customer %d is waiting (chairs occupied: %d)\n",
		customerID, len(sb.waitingRoom))

	// Wake up barber if sleeping
	if sb.barberSleeping {
		fmt.Printf("👋 Customer %d woke up the barber\n", customerID)
		sb.barberWake.Signal()
	}

	// Wait for haircut to be completed
	// Customer waits until their turn comes and haircut is done
	for {
		// Check if this customer is being served or shop is closed
		if !sb.isOpen {
			fmt.Printf("🚪 Customer %d left due to shop closure\n", customerID)
			return
		}

		// Check if customer is no longer in waiting room (being served or done)
		found := false
		for _, id := range sb.waitingRoom {
			if id == customerID {
				found = true
				break
			}
		}

		// If not in waiting room and barber is not busy with this customer
		if !found && !sb.barberBusy {
			fmt.Printf("😊 Customer %d got haircut and left happy\n", customerID)
			return
		}

		// Wait for condition change
		sb.customerWait.Wait()
	}
}

func (sb *SleepingBarber) closeShop() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	fmt.Println("\n🕐 Simulation time ended - closing shop")
	sb.isOpen = false

	// Wake up barber and all waiting customers
	sb.barberWake.Broadcast()
	sb.customerWait.Broadcast()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	waitingChairs := 3
	fmt.Printf("🏪 Opening the Sleeping Barber Shop with %d waiting chairs\n", waitingChairs)

	shop := NewSleepingBarber(waitingChairs)

	// Start the barber
	shop.wg.Add(1)
	go shop.barberWork()

	// Generate customers at random intervals
	customerCount := 0
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(20 * time.Second)

	for {
		select {
		case <-ticker.C:
			// Random chance of customer arriving (70% probability)
			if rand.Float32() < 0.7 {
				customerCount++
				shop.wg.Add(1)
				go shop.customerArrival(customerCount)
			}

		case <-timeout:
			// Close the shop
			shop.closeShop()

			// Wait for all goroutines to finish
			shop.wg.Wait()

			fmt.Println("👋 Barber shop closed")
			return
		}
	}
}
