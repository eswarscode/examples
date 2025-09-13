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