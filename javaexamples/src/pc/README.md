# Producer-Consumer Problem with Locks and Condition Variables

A Java implementation of the classic producer-consumer concurrency problem using `ReentrantLock` and `Condition` variables.

## Problem Description

The producer-consumer problem involves:
- **Producers**: Generate items and place them in a shared buffer
- **Consumers**: Remove items from the shared buffer and process them
- **Bounded Buffer**: Fixed-size buffer that can become full or empty
- **Synchronization**: Coordinate access to prevent race conditions and deadlocks

## Solution Overview

This implementation uses Java's explicit locking mechanisms:

### Key Components

- **ReentrantLock**: Provides exclusive access to the shared buffer
- **Condition Variables**:
  - `notFull`: Producers wait when buffer is full
  - `notEmpty`: Consumers wait when buffer is empty
- **SharedBuffer**: Thread-safe bounded buffer with configurable capacity
- **Producer Threads**: Create items and add them to the buffer
- **Consumer Threads**: Remove and process items from the buffer

### Synchronization Strategy

1. **Producer Logic**:
   ```java
   lock.lock()
   while (buffer is full)
       notFull.await()
   add item to buffer
   notEmpty.signalAll()
   lock.unlock()
   ```

2. **Consumer Logic**:
   ```java
   lock.lock()
   while (buffer is empty)
       notEmpty.await()
   remove item from buffer
   notFull.signalAll()
   lock.unlock()
   ```

### Features

- **Multiple Producers**: 3 producer threads each creating 3 items
- **Multiple Consumers**: 4 consumer threads each consuming 2 items
- **Bounded Buffer**: Capacity of 5 items with overflow/underflow protection
- **Random Delays**: Simulates realistic production/consumption rates
- **Detailed Logging**: Shows buffer state and thread operations
- **Graceful Shutdown**: All threads complete their work and terminate properly

## Running the Code

```bash
# Compile
javac ProducerConsumer.java

# Run
java ProducerConsumer
```

## Expected Output

```
Starting Producer-Consumer Demo
Buffer capacity: 5
Producers: 3 (each producing 3 items)
Consumers: 4 (each consuming 2 items)
Total items to produce: 9
Total items to consume: 8

Producer-1 produced: 101 [Buffer size: 1]
Producer-2 produced: 201 [Buffer size: 2]
Consumer-1 consumed: 101 [Buffer size: 1]
Producer-3 produced: 301 [Buffer size: 2]
Consumer-3 waiting - buffer empty
Producer-1 produced: 102 [Buffer size: 1]
Consumer-3 consumed: 201 [Buffer size: 0]
...
Demo completed. Final buffer size: 1
```

## Algorithm Analysis

### Thread Safety
- **Mutual Exclusion**: Only one thread can access buffer at a time
- **No Race Conditions**: Lock protects all shared state modifications
- **Deadlock Prevention**: Proper lock ordering and condition usage

### Efficiency
- **Blocking Operations**: Threads sleep when conditions aren't met (no busy waiting)
- **Condition Signaling**: `signalAll()` wakes appropriate waiting threads
- **Resource Utilization**: Multiple producers/consumers work concurrently when possible

### Configuration

You can modify these constants in the main method:
```java
final int BUFFER_SIZE = 5;           // Buffer capacity
final int ITEMS_PER_PRODUCER = 3;    // Items each producer creates
final int ITEMS_PER_CONSUMER = 2;    // Items each consumer processes
final int NUM_PRODUCERS = 3;         // Number of producer threads
final int NUM_CONSUMERS = 4;         // Number of consumer threads
```

## Key Learning Points

1. **ReentrantLock vs synchronized**: Explicit lock control and condition variables
2. **Condition Variables**: More precise thread coordination than `wait()/notify()`
3. **Proper Resource Management**: try-finally blocks ensure lock release
4. **Producer-Consumer Patterns**: Foundation for many concurrent systems
5. **Buffer Management**: Handling bounded resources in multi-threaded environments