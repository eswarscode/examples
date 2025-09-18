import java.util.LinkedList;
import java.util.Queue;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

class SharedBuffer {
    private final Queue<Integer> buffer;
    private final int capacity;
    private final ReentrantLock lock;
    private final Condition notFull;
    private final Condition notEmpty;

    public SharedBuffer(int capacity) {
        this.capacity = capacity;
        this.buffer = new LinkedList<>();
        this.lock = new ReentrantLock();
        this.notFull = lock.newCondition();
        this.notEmpty = lock.newCondition();
    }

    public void produce(int item) throws InterruptedException {
        lock.lock();
        try {
            while (buffer.size() == capacity) {
                System.out.println(Thread.currentThread().getName() + " waiting - buffer full");
                notFull.await();
            }

            buffer.offer(item);
            System.out.println(Thread.currentThread().getName() + " produced: " + item +
                             " [Buffer size: " + buffer.size() + "]");

            notEmpty.signalAll();
        } finally {
            lock.unlock();
        }
    }

    public int consume() throws InterruptedException {
        lock.lock();
        try {
            while (buffer.isEmpty()) {
                System.out.println(Thread.currentThread().getName() + " waiting - buffer empty");
                notEmpty.await();
            }

            int item = buffer.poll();
            System.out.println(Thread.currentThread().getName() + " consumed: " + item +
                             " [Buffer size: " + buffer.size() + "]");

            notFull.signalAll();
            return item;
        } finally {
            lock.unlock();
        }
    }

    public int size() {
        lock.lock();
        try {
            return buffer.size();
        } finally {
            lock.unlock();
        }
    }
}

class Producer implements Runnable {
    private final SharedBuffer buffer;
    private final int itemCount;
    private final int producerId;

    public Producer(SharedBuffer buffer, int itemCount, int producerId) {
        this.buffer = buffer;
        this.itemCount = itemCount;
        this.producerId = producerId;
    }

    @Override
    public void run() {
        try {
            for (int i = 1; i <= itemCount; i++) {
                int item = producerId * 100 + i;
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

class Consumer implements Runnable {
    private final SharedBuffer buffer;
    private final int itemCount;

    public Consumer(SharedBuffer buffer, int itemCount) {
        this.buffer = buffer;
        this.itemCount = itemCount;
    }

    @Override
    public void run() {
        try {
            for (int i = 0; i < itemCount; i++) {
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

public class ProducerConsumer {
    public static void main(String[] args) throws InterruptedException {
        final int BUFFER_SIZE = 5;
        final int ITEMS_PER_PRODUCER = 3;
        final int ITEMS_PER_CONSUMER = 2;
        final int NUM_PRODUCERS = 3;
        final int NUM_CONSUMERS = 4;

        SharedBuffer buffer = new SharedBuffer(BUFFER_SIZE);

        Thread[] producers = new Thread[NUM_PRODUCERS];
        Thread[] consumers = new Thread[NUM_CONSUMERS];

        System.out.println("Starting Producer-Consumer Demo");
        System.out.println("Buffer capacity: " + BUFFER_SIZE);
        System.out.println("Producers: " + NUM_PRODUCERS + " (each producing " + ITEMS_PER_PRODUCER + " items)");
        System.out.println("Consumers: " + NUM_CONSUMERS + " (each consuming " + ITEMS_PER_CONSUMER + " items)");
        System.out.println("Total items to produce: " + (NUM_PRODUCERS * ITEMS_PER_PRODUCER));
        System.out.println("Total items to consume: " + (NUM_CONSUMERS * ITEMS_PER_CONSUMER));
        System.out.println();

        for (int i = 0; i < NUM_PRODUCERS; i++) {
            producers[i] = new Thread(new Producer(buffer, ITEMS_PER_PRODUCER, i + 1));
            producers[i].setName("Producer-" + (i + 1));
        }

        for (int i = 0; i < NUM_CONSUMERS; i++) {
            consumers[i] = new Thread(new Consumer(buffer, ITEMS_PER_CONSUMER));
            consumers[i].setName("Consumer-" + (i + 1));
        }

        for (Thread producer : producers) {
            producer.start();
        }

        for (Thread consumer : consumers) {
            consumer.start();
        }

        for (Thread producer : producers) {
            producer.join();
        }

        for (Thread consumer : consumers) {
            consumer.join();
        }

        System.out.println("\nDemo completed. Final buffer size: " + buffer.size());
    }
}