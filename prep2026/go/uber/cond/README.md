# Uber Ride — Go (locks + condition variables only)

The Uber Ride puzzle solved with **only a `sync.Mutex` and two `sync.Cond`
variables** — no semaphores and no barrier. Compare with the semaphore version
in `../uber_ride.go`.

## The problem

A car may depart only with exactly 4 riders in a balanced combination:

- 4 Democrats, or
- 4 Republicans, or
- 2 Democrats + 2 Republicans.

A lone minority (1+3 / 3+1) is never allowed. All 4 must be seated before the
car departs, and one of them drives.

## Running

```sh
go run uber_ride.go
go run -race uber_ride.go
```

8 Democrats + 8 Republicans arrive interleaved; output is 4 rides, each showing
4 `seated` lines (a legal group) followed by a departure.

## Design

One lock guards everything. State:

| Field | Role |
|---|---|
| `waitingD` / `waitingR` | riders of each party currently waiting |
| `rideBusy` | true while a group is boarding — serializes group formation |
| `demTickets` / `repTickets` | seats granted to each party for the current group |
| `boarded` | how many of the current group have taken their seat |
| `cDem` / `cRep` (`sync.Cond`) | waiting riders of each party block here for a seat |

### Flow

**`tryForm()`** (called under the lock, on every arrival and after every
departure): if no ride is in progress and a legal group is available, it
subtracts the riders from the waiting counts, sets `demTickets`/`repTickets` to
the legal ratio (4+0 or 2+2), marks `rideBusy`, resets `boarded`, and
`Broadcast()`s both conditions.

**Rider** (Democrat; Republican is symmetric):
1. Lock, `waitingD++`, call `tryForm()`.
2. `for demTickets == 0 { cDem.Wait() }` — wait until granted a seat.
3. `demTickets--`, then `board()`.

**`board()`** (under the lock): print `seated`, `boarded++`. If this is the 4th
rider, it is the **driver**: print the departure, clear `rideBusy`, and call
`tryForm()` to assemble the next group. Unlock on return.

### Why no barrier is needed

Boarding happens entirely under the lock, so the four riders board one at a
time in some serial order. The **4th to board is by definition the last**, so it
can safely drive immediately — every earlier rider has already printed `seated`.
That ordering is exactly the "all seated before departure" guarantee a barrier
would provide, obtained for free from the monitor.

### Why it's correct

- **Only legal groups form.** `tryForm()` only ever sets tickets to 4+0 or 2+2,
  and only when the waiting counts actually support it (`waitingD >= 4`, or
  `waitingD >= 2 && waitingR >= 2`). So tickets never exceed available riders and
  a 1+3 split is impossible.
- **No overlapping groups.** `rideBusy` stays true from formation until the
  driver departs, so `tryForm()` can't start a second group that would corrupt
  the shared `boarded` counter.
- **No lost or spurious wake-ups.** Every `Wait()` is inside a `for` loop that
  re-checks the ticket count, so broadcasting to all waiters (some of whom won't
  get a seat this round) is safe.
- **Liveness / chaining.** `tryForm()` runs both when a rider arrives and when a
  ride departs, so remaining waiters are always re-evaluated — no group is left
  un-formed when the riders for it are present.
- **No data races.** `go run -race` is clean; all shared state is touched only
  under the lock.

### Termination note

8 + 8 (both even and ≥ 4) always drains to zero because every departure removes
(4,0), (0,4), or (2,2). An unbalanced supply (e.g. 3 Democrats, 0 Republicans)
could legally strand riders forever — inherent to the problem, not a bug.

### vs. the semaphore version

The semaphore version (`../uber_ride.go`) needs three counting semaphores plus a
separate reusable barrier, and the captain holds a mutex across the whole group
lifetime. This version collapses all of that into one monitor: the condition
variables replace the queue semaphores, the `rideBusy` flag replaces the
captain-holds-mutex trick, and the serialized boarding replaces the barrier.
