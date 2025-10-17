package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Semaphore implementation using channels
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
		// Channel is empty, nothing to release
	}
}

func (s *Semaphore) AvailablePermits() int {
	return cap(s.ch) - len(s.ch)
}

// SemaphoreBuffer represents a bounded buffer using semaphores
type SemaphoreBuffer struct {
	buffer []int
	mutex  *Semaphore // Binary semaphore for mutual exclusion
	empty  *Semaphore // Tracks empty slots
	full   *Semaphore // Tracks full slots
}

func NewSemaphoreBuffer(capacity int) *SemaphoreBuffer {
	return &SemaphoreBuffer{
		buffer: make([]int, 0, capacity),
		mutex:  NewSemaphore(1),           // Binary semaphore
		empty:  NewSemaphore(capacity),    // Initially all slots empty
		full:   NewSemaphore(0),           // Initially no items
	}
}

func (sb *SemaphoreBuffer) Produce(item int, producerID int) {
	sb.empty.Acquire() // Wait for empty slot
	sb.mutex.Acquire() // Enter critical section

	sb.buffer = append(sb.buffer, item)
	fmt.Printf("Producer-%d produced: %d [Buffer size: %d]\n", producerID, item, len(sb.buffer))

	sb.mutex.Release() // Exit critical section
	sb.full.Release()  // Signal item available
}

func (sb *SemaphoreBuffer) Consume(consumerID int) int {
	sb.full.Acquire()  // Wait for available item
	sb.mutex.Acquire() // Enter critical section

	item := sb.buffer[0]
	sb.buffer = sb.buffer[1:]
	fmt.Printf("Consumer-%d consumed: %d [Buffer size: %d]\n", consumerID, item, len(sb.buffer))

	sb.mutex.Release() // Exit critical section
	sb.empty.Release() // Signal slot available
	return item
}

func (sb *SemaphoreBuffer) Size() int {
	sb.mutex.Acquire()
	defer sb.mutex.Release()
	return len(sb.buffer)
}

func (sb *SemaphoreBuffer) PrintSemaphoreState() {
	fmt.Printf("Semaphore state - Empty permits: %d, Full permits: %d, Mutex permits: %d\n",
		sb.empty.AvailablePermits(), sb.full.AvailablePermits(), sb.mutex.AvailablePermits())
}

func producer(ctx context.Context, buffer *SemaphoreBuffer, producerID, itemCount int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= itemCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("Producer-%d interrupted\n", producerID)
			return
		default:
			item := producerID*100 + i

			// Check if buffer might be full
			if buffer.empty.AvailablePermits() == 0 {
				fmt.Printf("Producer-%d waiting - buffer full\n", producerID)
			}

			buffer.Produce(item, producerID)

			// Random delay
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		}
	}
	fmt.Printf("Producer-%d finished producing\n", producerID)
}

func consumer(ctx context.Context, buffer *SemaphoreBuffer, consumerID, itemCount int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < itemCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("Consumer-%d interrupted\n", consumerID)
			return
		default:
			// Check if buffer might be empty
			if buffer.full.AvailablePermits() == 0 {
				fmt.Printf("Consumer-%d waiting - buffer empty\n", consumerID)
			}

			buffer.Consume(consumerID)

			// Random delay
			time.Sleep(time.Duration(150+rand.Intn(250)) * time.Millisecond)
		}
	}
	fmt.Printf("Consumer-%d finished consuming\n", consumerID)
}

func main() {
	const (
		bufferSize       = 5
		itemsPerProducer = 3
		itemsPerConsumer = 2
		numProducers     = 3
		numConsumers     = 4
	)

	buffer := NewSemaphoreBuffer(bufferSize)
	var wg sync.WaitGroup

	fmt.Println("Starting Producer-Consumer Demo with Semaphores (Go)")
	fmt.Printf("Buffer capacity: %d\n", bufferSize)
	fmt.Printf("Producers: %d (each producing %d items)\n", numProducers, itemsPerProducer)
	fmt.Printf("Consumers: %d (each consuming %d items)\n", numConsumers, itemsPerConsumer)
	fmt.Printf("Total items to produce: %d\n", numProducers*itemsPerProducer)
	fmt.Printf("Total items to consume: %d\n", numConsumers*itemsPerConsumer)
	fmt.Println()

	buffer.PrintSemaphoreState()
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start producers
	for i := 1; i <= numProducers; i++ {
		wg.Add(1)
		go producer(ctx, buffer, i, itemsPerProducer, &wg)
	}

	// Start consumers
	for i := 1; i <= numConsumers; i++ {
		wg.Add(1)
		go consumer(ctx, buffer, i, itemsPerConsumer, &wg)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Println("\nDemo completed.")
	fmt.Printf("Final buffer size: %d\n", buffer.Size())
	buffer.PrintSemaphoreState()
}