import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

/**
 * The Uber Ride problem, solved with ONLY a {@link ReentrantLock} and
 * {@link Condition} variables (no semaphores, no barrier).
 *
 * A car may depart only with exactly 4 riders in a balanced combination:
 * 4 Democrats, 4 Republicans, or 2 Democrats + 2 Republicans. A lone minority
 * (1+3 / 3+1) is never allowed. All 4 must be seated before the car departs,
 * and one of them drives.
 *
 * Key idea: a single lock protects all state. Two conditions let waiting riders
 * of each party block until they're granted a seat ("ticket"). Group formation
 * is serialized by a rideBusy flag, so no illegal or overlapping group can
 * form. The "all seated before departure" rule needs no barrier: boarding
 * happens under the lock, so the 4th rider to board is trivially the last — it
 * becomes the driver and departs.
 */
public class UberRide {

    private final ReentrantLock lock = new ReentrantLock();
    private final Condition cDem = lock.newCondition(); // Democrats wait for a seat
    private final Condition cRep = lock.newCondition(); // Republicans wait for a seat

    private int waitingD = 0; // Democrats currently waiting
    private int waitingR = 0; // Republicans currently waiting

    private boolean rideBusy = false; // a group is currently boarding
    private int demTickets = 0;       // seats granted to Democrats for the group
    private int repTickets = 0;       // seats granted to Republicans for the group
    private int boarded = 0;          // how many of the group have taken their seat

    private int rideId = 0;

    /** Start a new group if a legal one can be assembled. Caller holds the lock. */
    private void tryForm() {
        if (rideBusy) {
            return;
        }
        if (waitingD >= 4) {
            waitingD -= 4;
            demTickets = 4;
        } else if (waitingD >= 2 && waitingR >= 2) {
            waitingD -= 2;
            waitingR -= 2;
            demTickets = 2;
            repTickets = 2;
        } else {
            return; // not enough riders for any legal group yet
        }
        rideBusy = true;
        boarded = 0;
        // Wake everyone; each rider re-checks its ticket count under the lock.
        cDem.signalAll();
        cRep.signalAll();
    }

    /** Take a seat and, if this is the 4th rider, drive. Caller holds the lock. */
    private void board(String party, int id) {
        System.out.printf("%s %d: seated.%n", party, id);
        boarded++;
        if (boarded == 4) {
            rideId++;
            System.out.printf(">>> Ride %d departing — driver: %s %d. <<<%n", rideId, party, id);
            rideBusy = false;
            tryForm(); // immediately try to assemble the next group
        }
    }

    public void democrat(int id) throws InterruptedException {
        lock.lock();
        try {
            waitingD++;
            tryForm();
            while (demTickets == 0) { // wait until granted a seat
                cDem.await();
            }
            demTickets--;
            board("Democrat", id);
        } finally {
            lock.unlock();
        }
    }

    public void republican(int id) throws InterruptedException {
        lock.lock();
        try {
            waitingR++;
            tryForm();
            while (repTickets == 0) {
                cRep.await();
            }
            repTickets--;
            board("Republican", id);
        } finally {
            lock.unlock();
        }
    }

    public static void main(String[] args) throws InterruptedException {
        final int dems = 8, reps = 8; // both even and >= 4 so every rider departs
        UberRide stop = new UberRide();

        Thread[] threads = new Thread[dems + reps];
        int t = 0;
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
                dems + reps, stop.rideId);
    }

    @FunctionalInterface
    private interface Rider {
        void ride() throws InterruptedException;
    }

    private static void runRider(Rider r) {
        try {
            r.ride();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
