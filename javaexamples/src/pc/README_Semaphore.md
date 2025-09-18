# Producer-Consumer Problem with Semaphores

A Java implementation of the classic producer-consumer concurrency problem using `Semaphore` synchronization primitives.

## Problem Description

The producer-consumer problem involves coordinating access to a shared bounded buffer between producer and consumer threads. This implementation uses semaphores to manage synchronization without explicit locks and condition variables.

## Solution Overview

This implementation uses three semaphores for coordination:

### Semaphore Strategy

1. **Empty Semaphore**: Tracks available empty slots in the buffer
   - Initial permits: `BUFFER_SIZE` (all slots are empty)
   - Producers acquire before producing (wait for empty slot)
   - Consumers release after consuming (free up a slot)

2. **Full Semaphore**: Tracks available items in the buffer
   - Initial permits: `0` (no items available)
   - Consumers acquire before consuming (wait for available item)
   - Producers release after producing (signal item available)

3. **Mutex Semaphore**: Provides mutual exclusion for buffer access
   - Initial permits: `1` (binary semaphore)
   - Both producers and consumers acquire/release for critical section

### Algorithm Flow

#### Producer Algorithm:
```java
empty.acquire()     // Wait for empty slot
mutex.acquire()     // Enter critical section
// Add item to buffer
mutex.release()     // Exit critical section
full.release()      // Signal item available
```

#### Consumer Algorithm:
```java
full.acquire()      // Wait for available item
mutex.acquire()     // Enter critical section
// Remove item from buffer
mutex.release()     // Exit critical section
empty.release()     // Signal slot available
```

## Key Features

- **Three-Semaphore Pattern**: Classic textbook implementation
- **No Busy Waiting**: Threads block on semaphore operations
- **Deadlock-Free**: Proper acquisition order prevents deadlocks
- **Scalable**: Supports multiple producers and consumers
- **Resource Tracking**: Semaphore permits track buffer state
- **State Monitoring**: Real-time semaphore permit display

## Running the Code

```bash
# Compile
javac ProducerConsumerSemaphore.java

# Run
java ProducerConsumerSemaphore
```

## Expected Output

```
Starting Producer-Consumer Demo with Semaphores
Buffer capacity: 5
Producers: 3 (each producing 3 items)
Consumers: 4 (each consuming 2 items)
Total items to produce: 9
Total items to consume: 8

Semaphore state - Empty permits: 5, Full permits: 0, Mutex permits: 1

Producer-1 produced: 101 [Buffer size: 1]
Producer-2 produced: 201 [Buffer size: 2]
Consumer-1 consumed: 101 [Buffer size: 1]
Producer-3 waiting - buffer full
Consumer-2 consumed: 201 [Buffer size: 0]
...
Demo completed.
Final buffer size: 1
Semaphore state - Empty permits: 4, Full permits: 1, Mutex permits: 1
```

## Algorithm Analysis

### Semaphore Benefits

1. **Simplicity**: No explicit condition variables or complex locking
2. **Counting**: Naturally tracks resource availability
3. **Blocking**: Built-in thread suspension when resources unavailable
4. **Fairness**: FIFO queuing for waiting threads (implementation dependent)

### Synchronization Properties

- **Mutual Exclusion**: Mutex semaphore ensures only one thread accesses buffer
- **Resource Counting**: Empty/full semaphores track buffer state
- **Deadlock Prevention**: Consistent acquisition order (empty/full → mutex)
- **Progress**: Guaranteed progress when resources become available

### Performance Characteristics

- **Low Overhead**: Simple acquire/release operations
- **Scalability**: Multiple threads can wait on different semaphores
- **Resource Efficiency**: No polling or busy waiting

## Comparison with Lock-Based Implementation

| Aspect | Semaphores | Locks + Conditions |
|--------|------------|-------------------|
| **Complexity** | Lower | Higher |
| **Resource Tracking** | Built-in counting | Manual state management |
| **Flexibility** | Limited | More control options |
| **Performance** | Comparable | Comparable |
| **Readability** | More intuitive | More explicit |

## Configuration

Modify these constants for different scenarios:
```java
final int BUFFER_SIZE = 5;           // Buffer capacity
final int ITEMS_PER_PRODUCER = 3;    // Items each producer creates
final int ITEMS_PER_CONSUMER = 2;    // Items each consumer processes
final int NUM_PRODUCERS = 3;         // Number of producer threads
final int NUM_CONSUMERS = 4;         // Number of consumer threads
```

## Key Learning Points

1. **Semaphore Patterns**: Understanding counting vs binary semaphores
2. **Resource Management**: Using semaphores to track available resources
3. **Synchronization Design**: Three-semaphore producer-consumer pattern
4. **Deadlock Avoidance**: Proper semaphore acquisition ordering
5. **Classic Algorithms**: Foundation pattern for many concurrent systems

## Common Pitfalls

1. **Wrong Acquisition Order**: Can lead to deadlocks
2. **Missing Releases**: Can cause resource leaks
3. **Race Conditions**: Forgetting mutex for critical sections
4. **Semaphore Initialization**: Incorrect initial permit counts

This semaphore-based implementation demonstrates a fundamental approach to solving the producer-consumer problem with clean, understandable synchronization primitives.