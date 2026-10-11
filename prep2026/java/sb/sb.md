Three semaphores + one lock coordinate the barber and customers:

| Primitive | Role |
|---|---|
| `customersReady` | Barber sleeps on this (`acquire`); a customer `release`s it to wake him. |
| `barberReady` | Customer waits on this until the barber invites them into the chair. |
| `haircutDone` | Customer waits on this until the cut is finished (prevents them leaving early). |
| `lock` + `waitingCount` | Guards the shared count so the "is a chair free?" check and the increment are atomic. |

The flow:
- **Barber** loops: `acquire(customersReady)` → sleeps when nobody's there → decrements `waitingCount` → `release(barberReady)` → cuts → `release(haircutDone)`.
- **Customer** takes the lock, and if `waitingCount < waitingChairs` sits down and increments; otherwise unlocks and leaves immediately (balking). A seated customer wakes the barber and blocks on `barberReady`, then `haircutDone`.

Key correctness points:
- The lock makes the full-shop check and the increment one atomic step — without it two customers could both see the last free chair.
- `haircutDone` keeps the customer present for the whole cut, so the handoff is a true rendezvous rather than fire-and-forget.
- `close()` flips `open` and releases the barber once more so he isn't stuck blocked on `customersReady` at shutdown.

The run output shows the expected behavior: the barber sleeps and wakes, up to 3 customers wait, and the rest (7, 8, 10) balk when the shop is full.
