# Sleeping Barber — Java

Two Java solutions to the classic Sleeping Barber synchronization problem.

| File | Primitive | Idiom |
|---|---|---|
| `sb/SleepingBarber.java` | `Semaphore` × 3 + `ReentrantLock` | classic semaphore solution |
| `cond/SleepingBarber.java` | `ReentrantLock` + `Condition` | lock + condition variables |

## The problem

A barber shop has one barber, one barber chair, and a waiting room with a fixed
number of chairs.

- If there are no customers, the barber **sleeps**.
- A customer **wakes** the barber if he's asleep.
- If the barber is busy and a waiting chair is free, the customer **waits**.
- If all waiting chairs are full, the customer **leaves** (this is called *balking*).

The challenge is to coordinate the barber and customers without lost wake-ups,
deadlock, or two customers claiming the last chair.

## Running

```sh
# semaphore version
cd sb && javac SleepingBarber.java && java SleepingBarber

# lock + condition-variable version
cd cond && javac SleepingBarber.java && java SleepingBarber
```

Both are configured for 3 waiting chairs and 10 customers arriving every 40ms
while each haircut takes 100ms — so the barber falls behind and some customers
balk.

---

## Condition-variable solution (`cond/SleepingBarber.java`)

This version uses one `ReentrantLock` to protect all shared state, plus
`Condition` variables obtained from that lock via `lock.newCondition()`. The
conditions replace the three semaphores a textbook solution would use.

### Shared state (all guarded by `lock`)

| Field | Meaning |
|---|---|
| `queue` (`ArrayDeque<Customer>`) | FIFO of seated, waiting customers. Its size is the number of occupied waiting chairs. |
| `waitingChairs` | Capacity of the waiting room (constant). |
| `open` | `false` once the shop is closing. |

### Condition variables

| Condition | Who awaits | Signaled when |
|---|---|---|
| `customerArrived` (on the shop) | the **barber**, when the queue is empty | a customer joins the queue, or the shop closes |
| `seatReady` (per **customer**) | the customer, until `inChair` is true | the barber takes them from the queue |
| `cutFinished` (per **customer**) | the customer, until `cutDone` is true | the barber finishes their haircut |

Each `Customer` carries its own `seatReady` / `cutFinished` conditions and two
boolean flags (`inChair`, `cutDone`). All of these conditions are created from
the **same** `ReentrantLock`, so awaiting one atomically releases the single
shared lock. This lets the barber hand back to a *specific* person rather than
"whoever wakes first."

### Flow

**Barber loop:**
1. `lock()`. `while (queue.isEmpty() && open) customerArrived.await();` — sleep while idle.
2. If closed and the queue is empty, break and go home.
3. `poll()` the front customer (FIFO), set `inChair = true`, `seatReady.signal()`, unlock.
4. Cut hair **outside the lock** (so new customers can still sit down).
5. `lock()`, set `cutDone = true`, `cutFinished.signal()`, unlock. Repeat.

**Customer (`visit`):**
1. `lock()`. If `queue.size() >= waitingChairs` or `!open` → **balk** and return (lock released in `finally`).
2. Otherwise `queue.add(this)` and `customerArrived.signal()` to wake the barber.
3. `while (!inChair) seatReady.await();` — wait to be invited into the chair.
4. `while (!cutDone) cutFinished.await();` — wait for the cut to finish.
5. Release the lock and leave.

**Shutdown (`close`):** set `open = false` and `customerArrived.signalAll()` so a
sleeping barber wakes, sees the shop is closed with an empty queue, and exits.

### Why it's correct

- **No lost wake-ups.** Every `await()` sits inside a `while` loop that
  re-checks its predicate. `Condition.await()` atomically releases the lock and
  re-acquires it on wake, so a signal arriving between the check and the await
  is never missed, and spurious wake-ups are harmless. (This is exactly why you
  always guard `await()` with `while`, never `if`.)
- **No race on the last chair.** The "is a chair free?" check and the
  `queue.add` happen under the same held lock, so two customers can never both
  claim the last seat.
- **FIFO fairness.** The `ArrayDeque` is a true first-in-first-out queue, so
  customers are served in arrival order — something raw semaphores do **not**
  guarantee. A single "next customer" field instead of a queue would be a bug:
  a later arrival would overwrite an earlier waiter, who would then block
  forever.
- **Haircut is a rendezvous.** `cutFinished` keeps the customer present for the
  whole cut, so the barber→customer handoff completes rather than being
  fire-and-forget.
- **Lock always released.** Each `visit()` and barber critical section uses
  `try { ... } finally { lock.unlock(); }`, so an exception can never leave the
  lock held.
- **Clean shutdown.** `signalAll()` on close guarantees the barber is not left
  blocked in `await()` forever.

### Semaphore version vs. condition-variable version

The semaphore version (`sb/SleepingBarber.java`) models each wait/wake edge as
its own counting semaphore (`customersReady`, `barberReady`, `haircutDone`) plus
a lock around the waiting count. The condition-variable version keeps the single
shared lock but expresses the same waits as predicates (`inChair`, `cutDone`,
queue non-empty) guarded by `while` loops. The condition-variable style makes
the shared state and its invariants explicit and maps directly onto how the
problem is solved with `wait/notify` or C `pthread_cond_t` — a common
interview ask. The semaphore style is terser when each handoff is a simple
one-to-one signal.
