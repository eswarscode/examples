# Sleeping Barber — Go

Two Go solutions to the classic Sleeping Barber synchronization problem.

| File | Primitive | Idiom |
|---|---|---|
| `sleeping_barber.go` | buffered channel | "share memory by communicating" |
| `cond/sleeping_barber.go` | `sync.Mutex` + `sync.Cond` | classic shared-state + condition variables |

## The problem

A barber shop has one barber, one barber chair, and a waiting room with a fixed
number of chairs.

- If there are no customers, the barber **sleeps**.
- A customer **wakes** the barber if he's asleep.
- If the barber is busy and a waiting chair is free, the customer **waits**.
- If all waiting chairs are full, the customer **leaves** (this is called *balking*).

The challenge is to coordinate the barber and the customers without lost
wake-ups, deadlock, or two customers claiming the last chair.

## Running

```sh
# channel version
go run sleeping_barber.go

# mutex + condition-variable version
cd cond && go run sleeping_barber.go

# with the race detector
go run -race sleeping_barber.go
```

Both are configured for 3 waiting chairs and 10 customers arriving every 40ms
while each haircut takes 100ms — so the barber falls behind and some customers
balk.

---

## Condition-variable solution (`cond/sleeping_barber.go`)

This version uses one `sync.Mutex` to protect all shared state, plus three
`sync.Cond` condition variables layered on that mutex. The conditions replace
the three semaphores a textbook solution would use.

### Shared state (all guarded by `shop.mu`)

| Field | Meaning |
|---|---|
| `queue []*customer` | FIFO of seated, waiting customers. Its length is the number of occupied waiting chairs. |
| `waitingChairs` | Capacity of the waiting room (constant). |
| `open` | `false` once the shop is closing. |

### Condition variables

| Cond | Who waits | Woken when |
|---|---|---|
| `customerArrived` (on `shop`) | the **barber**, when the queue is empty | a customer joins the queue, or the shop closes |
| `seatReady` (per **customer**) | the customer, until `inChair` is true | the barber takes them from the queue |
| `cutFinished` (per **customer**) | the customer, until `cutDone` is true | the barber finishes their haircut |

Each customer carries its own `seatReady` / `cutFinished` conditions and two
boolean flags (`inChair`, `cutDone`). This lets the barber hand back to a
*specific* person rather than "whoever wakes first."

### Flow

**Barber loop:**
1. Lock. `for len(queue) == 0 && open { customerArrived.Wait() }` — sleep while idle.
2. If closed and the queue is empty, unlock and go home.
3. Pop the front customer (FIFO), set `inChair = true`, `seatReady.Signal()`, unlock.
4. Cut hair **outside the lock** (so new customers can still sit down).
5. Lock, set `cutDone = true`, `cutFinished.Signal()`, unlock. Repeat.

**Customer (`visit`):**
1. Lock. If `len(queue) >= waitingChairs` or `!open` → unlock and **balk**.
2. Otherwise append to `queue`, `customerArrived.Signal()` to wake the barber.
3. `for !inChair { seatReady.Wait() }` — wait to be invited into the chair.
4. `for !cutDone { cutFinished.Wait() }` — wait for the cut to finish.
5. Unlock and leave.

**Shutdown (`close`):** set `open = false` and `customerArrived.Broadcast()` so
a sleeping barber wakes, sees the shop is closed with an empty queue, and exits.

### Why it's correct

- **No lost wake-ups.** Every `Wait()` sits inside a `for` loop that re-checks
  its predicate. `Cond.Wait()` atomically releases the mutex and re-acquires it
  on wake, so a signal that arrives between the check and the wait is never
  missed, and spurious wake-ups are harmless.
- **No race on the last chair.** The "is a chair free?" check and the
  `append` to `queue` happen under the same held mutex, so two customers can
  never both claim the last seat. (`go run -race` confirms no data races.)
- **FIFO fairness.** The queue is a true first-in-first-out slice, so customers
  are served in arrival order — something raw semaphores do **not** guarantee.
  This is why the FIFO slice matters: a single "next customer" slot would let a
  later arrival overwrite an earlier waiter, who would then block forever.
- **Haircut is a rendezvous.** `cutFinished` keeps the customer present for the
  whole cut, so the barber→customer handoff completes rather than being
  fire-and-forget.
- **Clean shutdown.** `Broadcast()` on close guarantees the barber is not left
  blocked in `Wait()` forever.

### Channel version vs. condition-variable version

The channel version (`sleeping_barber.go`) is shorter because the buffered
channel *is* the waiting room: its capacity enforces the chair limit, a
non-blocking `select` send models balking, and `range` over the channel makes
the barber sleep when it's empty. The condition-variable version is more
verbose but mirrors how the problem is solved in languages without channels
(C with pthreads, Java with `wait/notify`), and makes the shared state and its
invariants explicit — useful for interviews that specifically ask for locks and
condition variables.
