import java.util.concurrent.BrokenBarrierException;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.Semaphore;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * The Uber Ride problem.
 *
 * At the end of a political conference, Democrats and Republicans are leaving
 * in Ubers. To keep the peace, a car may only depart when it holds exactly 4
 * riders in one of these balanced combinations:
 *   - 4 Democrats, or
 *   - 4 Republicans, or
 *   - 2 Democrats + 2 Republicans.
 *
 * A lone minority (1+3 or 3+1) is never allowed. When a valid group forms, the
 * rider who completes it becomes the "captain" and drives; all 4 must be seated
 * before the car departs.
 *
 * Classic solution: a {@code mutex} semaphore guards the waiting counts; two
 * "queue" semaphores block riders until a valid group can form; and a reusable
 * {@link CyclicBarrier} makes all 4 riders rendezvous before the captain
 * drives. The captain holds the mutex from the moment the group forms until the
 * car departs, so no second group can form concurrently.
 */
public class UberRide {

    private final Semaphore mutex = new Semaphore(1); // guards the counts below
    private int democrats = 0;                        // Democrats currently waiting
    private int republicans = 0;                      // Republicans currently waiting

    private final Semaphore demsQueue = new Semaphore(0); // Democrats wait here
    private final Semaphore repsQueue = new Semaphore(0); // Republicans wait here
    private final CyclicBarrier barrier = new CyclicBarrier(4); // 4-rider rendezvous

    private final AtomicInteger rideId = new AtomicInteger();

    private void seated(String party, int id) {
        System.out.printf("%s %d: seated.%n", party, id);
    }

    private void drive(String party, int id) {
        int ride = rideId.incrementAndGet();
        System.out.printf(">>> Ride %d departing — captain: %s %d. <<<%n", ride, party, id);
    }

    public void democrat(int id) throws InterruptedException, BrokenBarrierException {
        mutex.acquire();
        boolean captain = false;
        democrats++;
        if (democrats == 4) {
            // Four Democrats present: release all four (3 waiting + me).
            demsQueue.release(4);
            democrats -= 4;
            captain = true;
        } else if (democrats == 2 && republicans >= 2) {
            // Two of each present: release 2 Democrats and 2 Republicans.
            demsQueue.release(2);
            democrats -= 2;
            repsQueue.release(2);
            republicans -= 2;
            captain = true;
        } else {
            // Can't form a group yet: release the mutex and wait.
            mutex.release();
        }

        demsQueue.acquire(); // captain's own permit, or wait to be released
        seated("Democrat", id);
        barrier.await();
        if (captain) {
            drive("Democrat", id);
            mutex.release(); // held since the group formed
        }
    }

    public void republican(int id) throws InterruptedException, BrokenBarrierException {
        mutex.acquire();
        boolean captain = false;
        republicans++;
        if (republicans == 4) {
            repsQueue.release(4);
            republicans -= 4;
            captain = true;
        } else if (republicans == 2 && democrats >= 2) {
            repsQueue.release(2);
            republicans -= 2;
            demsQueue.release(2);
            democrats -= 2;
            captain = true;
        } else {
            mutex.release();
        }

        repsQueue.acquire();
        seated("Republican", id);
        barrier.await();
        if (captain) {
            drive("Republican", id);
            mutex.release();
        }
    }

    public static void main(String[] args) throws InterruptedException {
        final int dems = 8, reps = 8; // both even and >= 4 so every rider departs
        UberRide stop = new UberRide();

        Thread[] threads = new Thread[dems + reps];
        int t = 0;
        // Interleave arrivals so groups form from a mix of parties.
        for (int i = 1; i <= dems || i <= reps; i++) {
            if (i <= dems) {
                final int id = i;
                threads[t++] = new Thread(() -> runRider(() -> stop.democrat(id)), "D-" + id);
            }
            if (i <= reps) {
                final int id = i;
                threads[t++] = new Thread(() -> runRider(() -> stop.republican(id)), "R-" + id);
            }
        }
        for (Thread th : threads) {
            th.start();
            Thread.sleep(5); // small stagger to vary arrival order
        }
        for (Thread th : threads) {
            th.join();
        }
        System.out.printf("All %d riders have departed in %d rides.%n",
                dems + reps, stop.rideId.get());
    }

    @FunctionalInterface
    private interface Rider {
        void ride() throws InterruptedException, BrokenBarrierException;
    }

    private static void runRider(Rider r) {
        try {
            r.ride();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        } catch (BrokenBarrierException e) {
            throw new RuntimeException(e);
        }
    }
}
