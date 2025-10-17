# Go Goroutine and Channel Patterns

This document shows the correct patterns for using goroutines with channels, addressing the race conditions and timing issues in the original code.

## Problem: Race Condition with Channel Assignment

### Incorrect Pattern

```go
var ip <-chan *user.User

go func() {
    ip = process(getUsers())  // Writing to ip in goroutine
}()
time.Sleep(time.Second * 10)  // Hope it's ready

go read(ip)  // Reading from ip in main - RACE!
```

**Issues**:
1. Data race on `ip` variable (write in one goroutine, read in another)
2. Relies on timing (sleep) instead of proper synchronization
3. Unnecessary goroutine wrapper around `process()`

## Solution 1: Return Channel Immediately, Write Asynchronously

The key is to put the goroutine **inside** the producer function, not outside.

```go
func main() {
    // Get channel immediately
    ip := process(getUsers())

    // Use WaitGroup for proper synchronization
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        read(ip)
    }()

    wg.Wait() // Wait for completion
    fmt.Println("end")
}

// Process returns channel immediately and populates it asynchronously
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, len(users))

    // Goroutine INSIDE process() - writes to channel asynchronously
    go func() {
        for _, u := range users {
            e := &user.User{}
            *e = *u
            results <- e
        }
        close(results) // Close when done writing
    }()

    return results // Return immediately
}

func read(data <-chan *user.User) {
    for user := range data {
        fmt.Println("received:", *user)
    }
}
```

**Benefits**:
- No race condition on variables
- Channel returned immediately for consumers to reference
- Proper synchronization with WaitGroup instead of sleep
- Producer goroutine runs independently

## Solution 2: Use Channel to Signal Completion

```go
func main() {
    ip := process(getUsers())
    done := make(chan bool)

    go func() {
        read(ip)
        done <- true  // Signal when done
    }()

    <-done  // Wait for completion signal
    fmt.Println("end")
}
```

## Solution 3: Simple Synchronous (No Goroutines)

If you don't need concurrency, keep it simple:

```go
func main() {
    ip := process(getUsers())
    read(ip)  // Just read directly
    fmt.Println("end")
}

func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, len(users))

    // No goroutine - write synchronously
    for _, u := range users {
        e := &user.User{}
        *e = *u
        results <- e
    }
    close(results)

    return results
}
```

## Key Principles

### 1. Goroutine Placement
- **Wrong**: Wrap the function call in a goroutine
  ```go
  go func() {
      ip = someFunction()
  }()
  ```
- **Right**: Put goroutine inside the function
  ```go
  func someFunction() <-chan T {
      ch := make(chan T)
      go func() {
          // async work
      }()
      return ch
  }
  ```

### 2. Synchronization
- **Wrong**: Use `time.Sleep()` to "hope" things are ready
- **Right**: Use proper synchronization primitives:
  - `sync.WaitGroup` for waiting on multiple goroutines
  - Channels for signaling completion
  - `select` for coordinating multiple channels

### 3. Channel Ownership
- Return the channel immediately so consumers can reference it
- The producer goroutine owns closing the channel
- Always close channels when done writing (prevents goroutine leaks)

### 4. Buffered vs Unbuffered Channels
- **Unbuffered** (`make(chan T)`): Sender blocks until receiver is ready
- **Buffered** (`make(chan T, n)`): Sender blocks only when buffer is full
- Size buffer appropriately (e.g., `len(users)`) to avoid blocking

## Complete Fixed Example

```go
package main

import (
    "fmt"
    "sync"
    user "cilium.com/examples/internal/pkg1"
)

func main() {
    fmt.Println("Hello, World!")

    populateUsers(getUsers())
    fmt.Println("Cache lookup:", getValue("Bob1"))

    // Correct pattern: get channel immediately
    ip := process(getUsers())

    // Use WaitGroup for synchronization
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        read(ip)
    }()

    wg.Wait()
    fmt.Println("end")
}

func getUsers() []*user.User {
    return []*user.User{
        user.NewUser("Bob1", 21),
        user.NewUser("Bob2", 22),
        user.NewUser("Bob3", 23),
        user.NewUser("Bob4", 24),
    }
}

func populateUsers(users []*user.User) {
    cache = map[string]*user.User{}
    for _, u := range users {
        cache[u.Name] = u
    }
}

var cache map[string]*user.User

func getValue(name string) *user.User {
    return cache[name]
}

func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, len(users))

    go func() {
        for _, u := range users {
            e := &user.User{}
            *e = *u
            results <- e
        }
        close(results)
    }()

    return results
}

func read(data <-chan *user.User) {
    for user := range data {
        fmt.Println("received:", *user)
    }
}
```

## Running with Race Detector

To detect race conditions in your code:

```bash
go run -race test.go
```

This will report any data races at runtime, helping you identify synchronization issues.
