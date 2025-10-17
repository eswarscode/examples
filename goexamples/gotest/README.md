# Go Pointer and Channel Bug Example

This repository demonstrates a common concurrency bug in Go when sending pointers to reused variables through channels.

## The Problem

When sending pointers through a channel in a loop, reusing the same variable causes all pointers to reference the same memory location.

### Bug Example 1: Reusing a Variable

```go
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, 5)
    var e user.User = user.User{}  // Single variable declared outside loop
    for _, u := range users {
        e = *(u)                    // Update same variable
        results <- &e               // Send pointer to SAME variable
    }
    close(results)
    return results
}
```

**Problem**: All channel values point to the same memory address. After the loop completes, that address contains the last user's data.

**Output**:
```
resd  {Bob4 24}
resd  {Bob4 24}
resd  {Bob4 24}
resd  {Bob4 24}
```

### Bug Example 2: Reassigning Still Doesn't Help

```go
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, 5)
    var e user.User = user.User{}
    for _, u := range users {
        e = user.User{}             // Reset the struct value
        e = *(u)
        results <- &e               // Still pointer to SAME variable
    }
    close(results)
    return results
}
```

**Problem**: Even though you're resetting `e`, `&e` always points to the same stack address.

**Output**: Still broken - all Bob4.

## The Solutions

### Solution 1: Declare Variable Inside Loop

```go
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, 5)
    for _, u := range users {
        e := *u                     // NEW variable each iteration
        results <- &e               // Pointer to unique variable
    }
    close(results)
    return results
}
```

### Solution 2: Allocate New Pointer Each Iteration

```go
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, 5)
    var e *user.User = &user.User{}
    for _, u := range users {
        e = &user.User{}            // NEW heap allocation each iteration
        *e = *u                     // Copy data
        results <- e                // Send unique pointer
    }
    close(results)
    return results
}
```

### Solution 3: Just Send the Original Pointer

```go
func process(users []*user.User) <-chan *user.User {
    results := make(chan *user.User, 5)
    for _, u := range users {
        results <- u                // Send the existing pointer
    }
    close(results)
    return results
}
```

**Correct Output**:
```
resd  {Bob1 21}
resd  {Bob2 22}
resd  {Bob3 23}
resd  {Bob4 24}
```

## Key Takeaway

When sending pointers through channels in a loop:
- **Don't**: Declare the variable outside the loop and reuse it
- **Do**: Create a new variable (or pointer) in each iteration
- **Remember**: In Go, the loop variable address stays the same across iterations

## Running the Example

```bash
go run test.go
```

The current implementation uses Solution 2, which properly creates a new pointer for each user in the loop.
