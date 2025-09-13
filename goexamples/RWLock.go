package main

import (
	"sync"
)

// CustomRWLock provides a basic read-write lock implementation.
type CustomRWLock struct {
	mu         sync.Mutex // The main mutex protecting the state.
	readers    int        // Number of active readers.
	writers    int        // 1 if writer is active, 0 otherwise.
	writerCond *sync.Cond // Condition for writers to wait for readers to finish.
	readerCond *sync.Cond // Condition for readers to wait for a writer to finish.
}

// NewCustomRWLock creates and returns a new CustomRWLock.
func NewCustomRWLock() *CustomRWLock {
	rw := &CustomRWLock{}
	rw.writerCond = sync.NewCond(&rw.mu)
	rw.readerCond = sync.NewCond(&rw.mu)
	return rw
}

// RLock blocks until a read lock is acquired.
func (rw *CustomRWLock) RLock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Wait if there is an active writer.
	for rw.writers > 0 {
		rw.readerCond.Wait()
	}
	rw.readers++
}

// RUnlock releases a read lock.
func (rw *CustomRWLock) RUnlock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.readers--

	// If this was the last reader, signal a waiting writer.
	if rw.readers == 0 {
		rw.writerCond.Signal()
	}
}

// Lock blocks until a write lock is acquired.
func (rw *CustomRWLock) Lock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Wait if there are active readers or another writer.
	for rw.readers > 0 || rw.writers > 0 {
		rw.writerCond.Wait()
	}
	rw.writers = 1
}

// Unlock releases a write lock.
func (rw *CustomRWLock) Unlock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.writers = 0

	// Wake all waiting readers or one waiting writer.
	// This implementation favors readers to prevent writer starvation.
	rw.readerCond.Broadcast()
	rw.writerCond.Signal()
}
