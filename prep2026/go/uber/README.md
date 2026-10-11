# Uber Ride — Go

A Go solution to the classic Uber Ride concurrency puzzle.

## The problem

At the end of a political conference, Democrats and Republicans are leaving in
Ubers. To keep the peace, a car may only depart when it holds **exactly 4**
riders in one of these balanced combinations:

- 4 Democrats, or
- 4 Republicans, or
- 2 Democrats + 2 Republicans.

A lone minority — **1+3 or 3+1 is never allowed**. When a valid group forms, the
rider who completes it becomes the **captain** and drives the car. All 4 riders
must be **seated** before the car departs.

The synchronization challenge: block arriving riders until a legal group can be
assembled, make exactly those 4 proceed together, and never let a partial or
illegal group form.

## Running

```sh
go run uber_ride.go
go run -race uber_ride.go   # with the race detector
```

Configured for 8 Democrats + 8 Republicans arriving interleaved. Output shows
each rider being seated and each ride departing with its captain; the final line
reports the total rides (always 4 here).

## Design

The solution follows the canonical semaphore recipe (from *The Little Book of
Semaphores*), with Go-native building blocks:

| Primitive | Type | Role |
|---|---|---|
| `mu` | `sync.Mutex` | the "mutex" semaphore; guards `democrats` / `republicans` counts |
| `demsQueue` | counting semaphore (buffered channel) | Democrats block here until released into a group |
| `repsQueue` | counting semaphore (buffered channel) | Republicans block here until released into a group |
| `barrier` | reusable barrier (mutex + `sync.Cond`) | all 4 group members rendezvous before departure |

Go has no built-in counting semaphore, so `semaphore` is a thin wrapper over a
buffered channel: `acquire()` receives, `release(n)` sends `n` times. The
`barrier` is a reusable phase barrier built from a mutex and a condition
variable (`CyclicBarrier` has no stdlib equivalent).

### Rider algorithm (Democrat; Republican is symmetric)

1. `mu.Lock()` and increment `democrats`.
2. Decide whether this arrival **completes** a legal group:
   - `democrats == 4` → release 4 permits on `demsQueue`, subtract 4, become **captain**.
   - `democrats == 2 && republicans >= 2` → release 2 on `demsQueue` **and** 2 on `repsQueue`, subtract 2 from each, become **captain**.
   - otherwise → **not** a captain: `mu.Unlock()` and fall through to wait.
3. `demsQueue.acquire()` — the captain consumes its own permit immediately; a
   non-captain blocks here until some future captain releases a permit for it.
4. `seated(...)`.
5. `barrier.await()` — all 4 riders meet here.
6. If captain: `drive(...)`, then `mu.Unlock()`.

### Why it's correct

- **Only legal groups form.** A group is created only in the two `captain`
  branches, which release permits in exactly the legal ratios (4+0, 0+4, or
  2+2). A 1+3 split is unreachable because no branch ever releases 1 or 3
  permits.
- **The captain holds the mutex across the whole group's lifetime** — from the
  moment the group forms until the car departs. So the counts can't be mutated
  by another arrival mid-group, and no second group can form until this one
  drives. This serializes group formation and keeps the single shared
  `barrier` used by exactly 4 goroutines at a time.
- **No lost wake-ups / no illegal early release.** A non-captain releases the
  mutex *before* blocking on its queue semaphore, so the captain can always make
  progress; the channel-backed semaphore buffers the released permits until the
  waiting goroutine receives them.
- **Everyone departs seated.** The barrier guarantees all 4 are seated before
  the captain calls `drive()`.
- **No data races.** `go run -race` is clean; every access to the shared counts
  happens under `mu`, and the ride counter uses `sync/atomic`.

### Termination note

The demo uses 8 + 8 (both even and ≥ 4). Because every removal is (4,0), (0,4),
or (2,2) and all riders eventually arrive, the waiting counts always drain to
zero — no rider is stranded. An odd or unbalanced supply (e.g. 3 Democrats, 0
Republicans) could legally leave riders waiting forever; that's inherent to the
problem, not a bug in the solution.
