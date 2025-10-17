package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// MutexBuffer represents a bounded buffer using mutex and condition variables
type MutexBuffer struct {
	buffer   []int
	capacity int
	mutex    sync.RWMutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
}

func NewMutexBuffer(capacity int) *MutexBuffer {
	mb := &MutexBuffer{
		buffer:   make([]int, 0, capacity),
		capacity: capacity,
	}
	mb.notFull = sync.NewCond(&mb.mutex)
	mb.notEmpty = sync.NewCond(&mb.mutex)
	return mb
}

func (mb *MutexBuffer) Produce(item int, producerID int) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// Wait while buffer is full
	for len(mb.buffer) == mb.capacity {
		fmt.Printf("Producer-%d waiting - buffer full\n", producerID)
		mb.notFull.Wait()
	}

	mb.buffer = append(mb.buffer, item)
	fmt.Printf("Producer-%d produced: %d [Buffer size: %d]\n", producerID, item, len(mb.buffer))

	// Signal that buffer is not empty
	mb.notEmpty.Signal()
}

func (mb *MutexBuffer) Consume(consumerID int) int {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// Wait while buffer is empty
	for len(mb.buffer) == 0 {
		fmt.Printf("Consumer-%d waiting - buffer empty\n", consumerID)
		mb.notEmpty.Wait()
	}

	item := mb.buffer[0]
	mb.buffer = mb.buffer[1:]
	fmt.Printf("Consumer-%d consumed: %d [Buffer size: %d]\n", consumerID, item, len(mb.buffer))

	// Signal that buffer is not full
	mb.notFull.Signal()
	return item
}

func (mb *MutexBuffer) Size() int {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	return len(mb.buffer)
}

func (mb *MutexBuffer) PrintBufferState() {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	fmt.Printf("Buffer state - Size: %d, Capacity: %d, Available slots: %d\n",
		len(mb.buffer), mb.capacity, mb.capacity-len(mb.buffer))
}

func producerMutex(ctx context.Context, buffer *MutexBuffer, producerID, itemCount int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= itemCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("Producer-%d interrupted\n", producerID)
			return
		default:
			item := producerID*100 + i
			buffer.Produce(item, producerID)

			// Random delay
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		}
	}
	fmt.Printf("Producer-%d finished producing\n", producerID)
}

func consumerMutex(ctx context.Context, buffer *MutexBuffer, consumerID, itemCount int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < itemCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("Consumer-%d interrupted\n", consumerID)
			return
		default:
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

	buffer := NewMutexBuffer(bufferSize)
	var wg sync.WaitGroup

	fmt.Println("Starting Producer-Consumer Demo with Mutex and Condition Variables (Go)")
	fmt.Printf("Buffer capacity: %d\n", bufferSize)
	fmt.Printf("Producers: %d (each producing %d items)\n", numProducers, itemsPerProducer)
	fmt.Printf("Consumers: %d (each consuming %d items)\n", numConsumers, itemsPerConsumer)
	fmt.Printf("Total items to produce: %d\n", numProducers*itemsPerProducer)
	fmt.Printf("Total items to consume: %d\n", numConsumers*itemsPerConsumer)
	fmt.Println()

	buffer.PrintBufferState()
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start producers
	for i := 1; i <= numProducers; i++ {
		wg.Add(1)
		go producerMutex(ctx, buffer, i, itemsPerProducer, &wg)
	}

	// Start consumers
	for i := 1; i <= numConsumers; i++ {
		wg.Add(1)
		go consumerMutex(ctx, buffer, i, itemsPerConsumer, &wg)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Println("\nDemo completed.")
	fmt.Printf("Final buffer size: %d\n", buffer.Size())
	buffer.PrintBufferState()
}