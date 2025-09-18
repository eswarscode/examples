# Java Examples

This repository contains various Java examples and implementations.

## Custom Scheduled Executor

A custom implementation of a scheduled executor with `scheduleAtFixedRate` functionality.

### Files
- `CustomScheduledExecutor.java` - The main executor implementation
- `ScheduledExecutorDemo.java` - Demonstration of usage

### Features
- **Fixed-rate scheduling**: Maintains consistent intervals between task executions
- **Thread pool management**: Configurable core pool size
- **Exception handling**: Tasks continue executing even if individual executions throw exceptions
- **Cancellation support**: Tasks can be cancelled via `ScheduledFuture.cancel()`
- **Graceful shutdown**: Proper cleanup of resources
- **Thread-safe**: Uses atomic operations for safe concurrent access

### Usage Example
```java
CustomScheduledExecutor executor = new CustomScheduledExecutor(2);

ScheduledFuture<?> future = executor.scheduleAtFixedRate(() -> {
    System.out.println("Task executed at: " + System.currentTimeMillis());
}, 1, 2, TimeUnit.SECONDS); // Initial delay: 1s, Period: 2s

// Cancel the task
future.cancel(true);

// Shutdown executor
executor.shutdown();
```

### Running the Demo
```bash
javac ScheduledExecutorDemo.java CustomScheduledExecutor.java
java ScheduledExecutorDemo
```

The demo will:
1. Create a scheduled task that runs every 2 seconds
2. Execute the task 5 times
3. Cancel the task and shutdown the executor

## Custom ThreadPool Executor

A custom implementation of a thread pool executor from scratch.

### Files
- `CustomThreadPoolExecutor.java` - The main thread pool implementation
- `ThreadPoolExecutorDemo.java` - Demonstration of usage

### Features
- **Core and maximum pool sizes**: Configurable thread pool sizing
- **Keep-alive time**: Excess threads terminate after idle timeout
- **Work queue**: Bounded or unbounded task queuing
- **Thread factory**: Custom thread creation
- **Rejection policy**: Handles task rejection when pool is saturated
- **Thread-safe operations**: Uses locks and atomic operations
- **Lifecycle management**: Proper startup and shutdown handling

### Key Components
- **Worker threads**: Internal worker class that processes tasks
- **Task queuing**: BlockingQueue for task management
- **Pool size management**: Dynamic thread creation and destruction
- **Interruption handling**: Graceful thread interruption during shutdown

### Usage Example
```java
CustomThreadPoolExecutor executor = new CustomThreadPoolExecutor(
    2,                              // core pool size
    4,                              // maximum pool size
    60L,                            // keep alive time
    TimeUnit.SECONDS,               // time unit
    new LinkedBlockingQueue<>(10)   // work queue
);

// Submit tasks
executor.execute(() -> {
    System.out.println("Task running on: " + Thread.currentThread().getName());
});

// Shutdown
executor.shutdown();
executor.awaitTermination(10, TimeUnit.SECONDS);
```

### Running the Demo
```bash
javac ThreadPoolExecutorDemo.java CustomThreadPoolExecutor.java
java ThreadPoolExecutorDemo
```

The demo will:
1. Create a thread pool with 2 core threads, 4 max threads
2. Submit 15 tasks to demonstrate pool expansion and queue usage
3. Show task rejection when pool and queue are full
4. Demonstrate graceful shutdown

## Uber Seating Problem

A Java implementation of the classic Uber Seating concurrency problem.

### Problem Description

An Uber ride can seat exactly 4 passengers with the following constraints:
- Either 4 Democrats
- Either 4 Republicans
- Or 2 Democrats and 2 Republicans

Passengers arrive concurrently and must wait until a valid seating arrangement is possible.

### Solution Overview

The solution uses Java concurrency primitives:

- **Semaphores**: Block passengers until valid groups form
- **ReentrantLock**: Protects shared counters
- **CyclicBarrier**: Ensures all 4 passengers are seated before ride starts
- **Ride Leader Pattern**: One passenger coordinates the group and starts the ride

### Algorithm

1. Passenger arrives and increments their party counter
2. Check if they complete a valid group:
   - 4 of same party: Release 3 waiting passengers of same party
   - 2+2 mixed: Release 1 from same party + 2 from other party
3. If no valid group, wait on party semaphore
4. All passengers wait at barrier until 4 are seated
5. Ride leader starts the drive

### Running the Code

```bash
# Compile
javac src/Seating.java

# Run
java -cp src Demonstration
```

### Expected Output

```
Democrat_1 seated
Republican_1 seated
Republican_2 seated
Democrat_2 seated
Uber Ride on Its wayyyy... with ride leader Thread-1
...
```

The test creates 10 Democrat and 14 Republican threads, demonstrating various valid seating combinations.