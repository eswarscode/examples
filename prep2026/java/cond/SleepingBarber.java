import java.util.ArrayDeque;
import java.util.Queue;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

/**
 * The Sleeping Barber problem implemented with a {@link ReentrantLock} and
 * {@link Condition} variables rather than semaphores.
 *
 * A barber shop has one barber, one barber chair, and a waiting room with a
 * fixed number of chairs.
 *   - If there are no customers, the barber sleeps (Condition.await).
 *   - A customer wakes the barber by signaling the condition.
 *   - If the barber is busy and a waiting chair is free, the customer waits.
 *   - If all waiting chairs are full, the customer leaves (balks).
 *
 * All shared state (the waiting-room queue, the per-customer handoff flags, and
 * the open flag) is protected by a single lock. The Conditions layered on that
 * lock replace the semaphores a classic solution would use.
 */
public class SleepingBarber {

    private final int waitingChairs;
    private final ReentrantLock lock = new ReentrantLock();

    // FIFO of seated, waiting customers; its size is the number of occupied
    // waiting chairs. Guarded by lock.
    private final Queue<Customer> queue = new ArrayDeque<>();

    // Barber awaits here when the queue is empty.
    private final Condition customerArrived = lock.newCondition();

    private boolean open = true;

    public SleepingBarber(int waitingChairs) {
        this.waitingChairs = waitingChairs;
    }

    /**
     * Carries its own handoff state so the barber hands back to a specific
     * person rather than "whoever." All fields guarded by the enclosing lock.
     */
    private class Customer {
        final int id;
        boolean inChair;  // barber has invited this customer into the chair
        boolean cutDone;  // barber has finished this customer's haircut

        // Customer awaits on seatReady until inChair, then on cutFinished
        // until cutDone. Both backed by the shared lock.
        final Condition seatReady = lock.newCondition();
        final Condition cutFinished = lock.newCondition();

        Customer(int id) {
            this.id = id;
        }

        /** Arrive, wait if there's room, get a haircut; or balk if full. */
        void visit() throws InterruptedException {
            lock.lock();
            try {
                if (queue.size() >= waitingChairs || !open) {
                    System.out.printf("Customer %d: shop full, leaving.%n", id);
                    return;
                }

                // Sit down and wake the barber if he's asleep.
                queue.add(this);
                System.out.printf("Customer %d: sitting down to wait (waiting=%d).%n", id, queue.size());
                customerArrived.signal();

                // Wait until the barber invites us into the chair.
                while (!inChair) {
                    seatReady.await();
                }
                System.out.printf("Customer %d: getting a haircut.%n", id);

                // Wait until the haircut is finished.
                while (!cutDone) {
                    cutFinished.await();
                }
            } finally {
                lock.unlock();
            }
            System.out.printf("Customer %d: leaving, haircut done.%n", id);
        }
    }

    /** The barber: sleep until a customer arrives, serve them (FIFO), repeat. */
    private class Barber implements Runnable {
        @Override
        public void run() {
            try {
                while (true) {
                    Customer c;
                    lock.lock();
                    try {
                        // Sleep while nobody is waiting and the shop is open.
                        while (queue.isEmpty() && open) {
                            customerArrived.await();
                        }
                        // Closed and nobody left to serve -> go home.
                        if (queue.isEmpty() && !open) {
                            break;
                        }
                        // Take the next waiting customer and invite them in.
                        c = queue.poll();
                        c.inChair = true;
                        c.seatReady.signal();
                    } finally {
                        lock.unlock();
                    }

                    // Cut hair outside the lock so others can still sit down.
                    cutHair(c);

                    // Signal the customer the haircut is complete.
                    lock.lock();
                    try {
                        c.cutDone = true;
                        c.cutFinished.signal();
                    } finally {
                        lock.unlock();
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
            System.out.println("Barber: no more customers, going home.");
        }

        private void cutHair(Customer c) throws InterruptedException {
            System.out.printf("Barber: cutting hair for customer %d...%n", c.id);
            Thread.sleep(100); // simulate the haircut
            System.out.printf("Barber: finished customer %d.%n", c.id);
        }
    }

    /** Shut the shop and wake the sleeping barber so he can exit. */
    public void close() {
        lock.lock();
        try {
            open = false;
            customerArrived.signalAll();
        } finally {
            lock.unlock();
        }
    }

    public static void main(String[] args) throws InterruptedException {
        SleepingBarber shop = new SleepingBarber(3); // 3 waiting chairs

        Thread barber = new Thread(shop.new Barber(), "barber");
        barber.start();

        // Customers arrive faster than the barber works, so some will balk.
        Thread[] customers = new Thread[10];
        for (int i = 0; i < customers.length; i++) {
            Customer c = shop.new Customer(i + 1);
            customers[i] = new Thread(() -> {
                try {
                    c.visit();
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            }, "customer-" + (i + 1));
            customers[i].start();
            Thread.sleep(40); // customers arrive every 40ms
        }

        // Let in-flight customers finish, close up, wait for the barber.
        for (Thread t : customers) {
            t.join();
        }
        shop.close();
        barber.join();
        System.out.println("Shop closed.");
    }
}
