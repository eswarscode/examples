import java.util.LinkedList;
import java.util.Queue;
import java.util.concurrent.Semaphore;
import java.util.concurrent.locks.ReentrantLock;

class SemaphoreBuffer {
    private final Queue<Integer> buffer;
    private final int capacity;
    final Semaphore empty;    // Tracks empty slots (producers can produce)
    final Semaphore full;     // Tracks full slots (consumers can consume)
    private final Semaphore mutex;    // Mutual exclusion for buffer access

    public SemaphoreBuffer(int capacity) {
        this.capacity = capacity;
        this.buffer = new LinkedList<>();
        this.empty = new Semaphore(capacity);  // Initially all slots are empty
        this.full = new Semaphore(0);          // Initially no items to consume
        this.mutex = new Semaphore(1);         // Binary semaphore for mutual exclusion
    }

    public void produce(int item) throws InterruptedException {
        empty.acquire();    // Wait for empty slot
        mutex.acquire();    // Enter critical section

        try {
            buffer.offer(item);
            System.out.println(Thread.currentThread().getName() + " produced: " + item +
                             " [Buffer size: " + buffer.size() + "]");
        } finally {
            mutex.release();    // Exit critical section
        }

        full.release();     // Signal that an item is available
    }

    public int consume() throws InterruptedException {
        full.acquire();     // Wait for available item
        mutex.acquire();    // Enter critical section

        int item;
        try {
            item = buffer.poll();
            System.out.println(Thread.currentThread().getName() + " consumed: " + item +
                             " [Buffer size: " + buffer.size() + "]");
        } finally {
            mutex.release();    // Exit critical section
        }

        empty.release();    // Signal that a slot is available
        return item;
    }

    public int size() throws InterruptedException {
        mutex.acquire();
        try {
            return buffer.size();
        } finally {
            mutex.release();
        }
    }

    public void printSemaphoreState() {
        System.out.println("Semaphore state - Empty permits: " + empty.availablePermits() +
                         ", Full permits: " + full.availablePermits() +
                         ", Mutex permits: " + mutex.availablePermits());
    }
}

class SemaphoreProducer implements Runnable {
    private final SemaphoreBuffer buffer;
    private final int itemCount;
    private final int producerId;

    public SemaphoreProducer(SemaphoreBuffer buffer, int itemCount, int producerId) {
        this.buffer = buffer;
        this.itemCount = itemCount;
        this.producerId = producerId;
    }

    @Override
    public void run() {
        try {
            for (int i = 1; i <= itemCount; i++) {
                int item = producerId * 100 + i;

                // Check if buffer might be full
                if (buffer.empty.availablePermits() == 0) {
                    System.out.println(Thread.currentThread().getName() + " waiting - buffer full");
                }

                buffer.produce(item);
                Thread.sleep(100 + (int)(Math.random() * 200)); // Random delay
            }
            System.out.println(Thread.currentThread().getName() + " finished producing");
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            System.out.println(Thread.currentThread().getName() + " interrupted");
        }
    }
}

class SemaphoreConsumer implements Runnable {
    private final SemaphoreBuffer buffer;
    private final int itemCount;

    public SemaphoreConsumer(SemaphoreBuffer buffer, int itemCount) {
        this.buffer = buffer;
        this.itemCount = itemCount;
    }

    @Override
    public void run() {
        try {
            for (int i = 0; i < itemCount; i++) {
                // Check if buffer might be empty
                if (buffer.full.availablePermits() == 0) {
                    System.out.println(Thread.currentThread().getName() + " waiting - buffer empty");
                }

                int item = buffer.consume();
                Thread.sleep(150 + (int)(Math.random() * 250)); // Random delay
            }
            System.out.println(Thread.currentThread().getName() + " finished consuming");
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            System.out.println(Thread.currentThread().getName() + " interrupted");
        }
    }
}

public class ProducerConsumerSemaphore {
    public static void main(String[] args) throws InterruptedException {
        final int BUFFER_SIZE = 5;
        final int ITEMS_PER_PRODUCER = 3;
        final int ITEMS_PER_CONSUMER = 2;
        final int NUM_PRODUCERS = 3;
        final int NUM_CONSUMERS = 4;

        SemaphoreBuffer buffer = new SemaphoreBuffer(BUFFER_SIZE);

        Thread[] producers = new Thread[NUM_PRODUCERS];
        Thread[] consumers = new Thread[NUM_CONSUMERS];

        System.out.println("Starting Producer-Consumer Demo with Semaphores");
        System.out.println("Buffer capacity: " + BUFFER_SIZE);
        System.out.println("Producers: " + NUM_PRODUCERS + " (each producing " + ITEMS_PER_PRODUCER + " items)");
        System.out.println("Consumers: " + NUM_CONSUMERS + " (each consuming " + ITEMS_PER_CONSUMER + " items)");
        System.out.println("Total items to produce: " + (NUM_PRODUCERS * ITEMS_PER_PRODUCER));
        System.out.println("Total items to consume: " + (NUM_CONSUMERS * ITEMS_PER_CONSUMER));
        System.out.println();

        buffer.printSemaphoreState();
        System.out.println();

        // Create producer threads
        for (int i = 0; i < NUM_PRODUCERS; i++) {
            producers[i] = new Thread(new SemaphoreProducer(buffer, ITEMS_PER_PRODUCER, i + 1));
            producers[i].setName("Producer-" + (i + 1));
        }

        // Create consumer threads
        for (int i = 0; i < NUM_CONSUMERS; i++) {
            consumers[i] = new Thread(new SemaphoreConsumer(buffer, ITEMS_PER_CONSUMER));
            consumers[i].setName("Consumer-" + (i + 1));
        }

        // Start all threads
        for (Thread producer : producers) {
            producer.start();
        }

        for (Thread consumer : consumers) {
            consumer.start();
        }

        // Wait for all threads to complete
        for (Thread producer : producers) {
            producer.join();
        }

        for (Thread consumer : consumers) {
            consumer.join();
        }

        System.out.println("\nDemo completed.");
        System.out.println("Final buffer size: " + buffer.size());
        buffer.printSemaphoreState();
    }
}