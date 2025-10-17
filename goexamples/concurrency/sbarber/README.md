# Sleeping Barber Problem - Go Solution

## Problem Description

The Sleeping Barber is a classic synchronization problem in computer science that demonstrates the challenges of coordinating concurrent processes.

**Scenario:**
- A barbershop has one barber and a waiting room with `n` chairs
- If no customers are present, the barber goes to sleep
- If a customer enters and all chairs are occupied, the customer leaves immediately
- If the barber is busy but chairs are available, the customer sits and waits
- If the barber is asleep, the customer wakes the barber up

## Solution Overview

This Go implementation uses goroutines and channels to solve the synchronization problem with proper sleeping behavior:

### Key Components

1. **Barbershop struct:**
   - `waitingRoom`: Buffered channel that acts as the waiting chairs (capacity = number of chairs)
   - `wakeUpBarber`: Buffered channel to wake the sleeping barber
   - `customerDone`: Unbuffered channel to signal haircut completion
   - `barberBusy`: Boolean flag to track if barber is currently cutting hair
   - `barberSleeping`: Boolean flag to track if barber is actually sleeping
   - `mutex`: Ensures thread-safe access to shared state

2. **Goroutines:**
   - **Barber goroutine**: Sleeps when no customers, wakes up to serve all waiting customers
   - **Customer goroutines**: Each customer runs in its own goroutine

### Synchronization Strategy

- **True Sleep State**: Barber actually blocks on `wakeUpBarber` channel when sleeping
- **Atomic Wake-up**: Only one customer can wake the barber (prevents race conditions)
- **Non-blocking Entry**: Customers use `select` with `default` to avoid blocking on full waiting room
- **Service Coordination**: Direct communication between barber and customers
- **Completion Signaling**: Unbuffered channel ensures customers wait for service completion

## How It Works

1. **Barber Sleep Cycle:**
   ```go
   // Barber sets sleeping state and blocks until woken
   bs.barberSleeping = true
   select {
   case <-bs.wakeUpBarber:  // Blocks until customer wakes barber
       // Process all customers in waiting room
   case <-time.After(10 * time.Second):  // Timeout to close shop
   }
   ```

2. **Customer Arrival:**
   ```go
   select {
   case bs.waitingRoom <- customerID:  // Try to take a seat (non-blocking)
       // Check if barber is sleeping and wake if needed
       if barberSleeping {
           bs.barberSleeping = false  // Atomic state change
           bs.wakeUpBarber <- struct{}{}  // Wake the barber
       }
   default:
       // All chairs occupied, customer leaves immediately
   }
   ```

3. **Race Condition Prevention:**
   - Only the first customer seeing `barberSleeping = true` can wake the barber
   - Subsequent customers see `barberSleeping = false` and just wait
   - Mutex ensures atomic check-and-set operations

4. **Service Completion:**
   - Barber signals completion via `customerDone` channel
   - Customer receives signal and leaves satisfied

## Steps to Run

### Prerequisites
- Go 1.18 or higher installed on your system

### Running the Program

1. **Clone or download the code:**
   ```bash
   # If you have the file locally
   cd /path/to/sleeping_barber.go
   ```

2. **Run the program:**
   ```bash
   go run sleeping_barber.go
   ```

3. **Expected Output:**
   ```
   Barbershop opened with 3 waiting chairs
   💤 Barber: Going to sleep...
   Customer 1: Entering barbershop
   Customer 1: Found barber sleeping, waking up the barber
   👋 Barber: Woken up!
   ✂️  Barber: Cutting hair for customer 1
   Customer 2: Entering barbershop
   Customer 2: Barber is busy, waiting in chair (1/3 chairs occupied)
   ...
   Customer 7: All chairs occupied, leaving without haircut
   ...
   💤 Barber: Going to sleep...
   🏠 Barber: No customers for too long, closing shop
   Barbershop closed
   ```

### Customization

You can modify the simulation parameters in the `main()` function:

```go
numChairs := 3      // Number of waiting chairs
numCustomers := 8   // Number of customers to simulate
```

### Understanding the Output

- **💤 Barber: Going to sleep** - Barber enters actual sleep state (blocks on channel)
- **Customer X: Entering barbershop** - Customer arrives
- **Customer X: Found barber sleeping, waking up the barber** - Customer wakes sleeping barber
- **👋 Barber: Woken up!** - Barber receives wake signal and becomes active
- **Customer X: Barber is busy, waiting in chair** - Customer sits in waiting room
- **Customer X: All chairs occupied, leaving without haircut** - Customer leaves immediately (no blocking)
- **✂️ Barber: Cutting hair for customer X** - Barber starts haircut
- **✅ Barber: Finished cutting hair for customer X** - Haircut complete
- **Customer X: Leaving with fresh haircut** - Satisfied customer leaves
- **🏠 Barber: No customers for too long, closing shop** - Barber times out and closes

## Key Features

- **True Sleep Implementation**: Barber actually blocks (sleeps) until woken by customers
- **Race Condition Prevention**: Atomic check-and-set prevents multiple customers from waking barber
- **Non-blocking Customer Logic**: Customers never block when waiting room is full
- **No Deadlocks**: Proper channel usage prevents goroutine deadlocks
- **Realistic Simulation**: Random timing mimics real-world scenarios
- **Resource Management**: Bounded waiting room prevents unlimited queuing
- **Clean Shutdown**: Barber times out and closes shop when no customers arrive

## Educational Value

This implementation demonstrates:
- **True blocking sleep states** using channel operations
- **Atomic state management** to prevent race conditions
- **Non-blocking operations** with select statements and default cases
- **Goroutine coordination** using multiple channels
- **Producer-consumer patterns** with bounded resources
- **Mutex-based synchronization** for shared state protection
- **Timeout handling** for graceful shutdown
- **Complex concurrent system design** following classic computer science problems

## Common Pitfalls Avoided

1. **Fake Sleep**: Using printf instead of actual blocking sleep state
2. **Race Conditions**: Multiple customers waking the barber simultaneously
3. **Blocking Customers**: Customers getting stuck when waiting room is full
4. **Deadlocks**: Improper channel coordination causing goroutine deadlocks
5. **Resource Leaks**: Goroutines not properly cleaning up on exit