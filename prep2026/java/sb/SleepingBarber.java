import java.util.concurrent.Semaphore;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.ReentrantLock;

/**
 * The Sleeping Barber problem.
 *
 * A barber shop has one barber, one barber chair, and a waiting room with a
 * fixed number of chairs.
 *   - If there are no customers, the barber sleeps in the chair.
 *   - A customer wakes the barber if he's asleep.
 *   - If the barber is busy and a waiting chair is free, the customer waits.
 *   - If all waiting chairs are full, the customer leaves.
 *
 * Synchronization is done with three semaphores plus a lock guarding the
 * shared count of waiting customers.
 */
public class SleepingBarber {

    private final int waitingChairs;

    // Counts customers currently waiting (not yet being served).
    private int waitingCount = 0;
    private final ReentrantLock lock = new ReentrantLock();

    // Barber waits on this until a customer is ready.
    private final Semaphore customersReady = new Semaphore(0);
    // Customer waits on this until the barber is ready to cut.
    private final Semaphore barberReady = new Semaphore(0);
    // Signals the customer that their haircut is finished.
    private final Semaphore haircutDone = new Semaphore(0);

    private volatile boolean open = true;

    public SleepingBarber(int waitingChairs) {
        this.waitingChairs = waitingChairs;
    }

    /** The barber thread: sleeps until woken, then cuts hair, forever. */
    private class Barber implements Runnable {
        @Override
        public void run() {
            try {
                while (open) {
                    // Sleep until a customer signals they're ready.
                    customersReady.acquire();
                    if (!open) {
                        break;
                    }

                    lock.lock();
                    try {
                        waitingCount--; // this customer leaves the waiting room
                    } finally {
                        lock.unlock();
                    }

                    // Invite the waiting customer into the barber chair.
                    barberReady.release();

                    cutHair();

                    // Tell the customer the haircut is done.
                    haircutDone.release();
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }

        private void cutHair() throws InterruptedException {
            System.out.println("Barber: cutting hair...");
            Thread.sleep(100); // simulate time to cut hair
            System.out.println("Barber: finished a haircut.");
        }
    }

    /** A customer thread: arrives, waits if possible, gets a haircut or leaves. */
    private class Customer implements Runnable {
        private final int id;

        Customer(int id) {
            this.id = id;
        }

        @Override
        public void run() {
            lock.lock();
            if (waitingCount < waitingChairs) {
                waitingCount++;
                System.out.printf("Customer %d: sitting down (waiting=%d).%n", id, waitingCount);
                lock.unlock();

                try {
                    // Wake the barber (or add to the count he'll see).
                    customersReady.release();
                    // Wait for the barber to be ready for me.
                    barberReady.acquire();
                    System.out.printf("Customer %d: getting a haircut.%n", id);
                    // Wait until the haircut is complete.
                    haircutDone.acquire();
                    System.out.printf("Customer %d: leaving, haircut done.%n", id);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            } else {
                // No free chair — leave immediately.
                lock.unlock();
                System.out.printf("Customer %d: shop full, leaving.%n", id);
            }
        }
    }

    public void close() {
        open = false;
        customersReady.release(); // wake the barber so he can exit
    }

    public static void main(String[] args) throws InterruptedException {
        SleepingBarber shop = new SleepingBarber(3); // 3 waiting chairs

        Thread barber = new Thread(shop.new Barber(), "barber");
        barber.start();

        // Simulate customers arriving at varying intervals.
        AtomicInteger idGen = new AtomicInteger(1);
        for (int i = 0; i < 10; i++) {
            int id = idGen.getAndIncrement();
            new Thread(shop.new Customer(id), "customer-" + id).start();
            Thread.sleep(40); // customers arrive faster than the barber works
        }

        // Let the shop run, then close up.
        Thread.sleep(2000);
        shop.close();
        barber.join();
        System.out.println("Shop closed.");
    }
}
