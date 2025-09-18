import java.util.Random;
import java.util.concurrent.*;

public class SleepingBarber {
    private final BlockingQueue<Integer> waitingRoom;
    private final CountDownLatch shutdownLatch;
    private final ExecutorService executor;
    private final Random random;
    private volatile boolean isOpen;

    public SleepingBarber(int waitingChairs) {
        this.waitingRoom = new ArrayBlockingQueue<>(waitingChairs);
        this.shutdownLatch = new CountDownLatch(1);
        this.executor = Executors.newCachedThreadPool();
        this.random = new Random();
        this.isOpen = true;
    }

    public void startBarber() {
        executor.submit(() -> {
            try {
                barberWork();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });
    }

    private void barberWork() throws InterruptedException {
        while (isOpen) {
            System.out.println("💤 Barber is sleeping...");

            try {
                // Wait for a customer (with timeout)
                Integer customerId = waitingRoom.poll(10, TimeUnit.SECONDS);

                if (customerId != null) {
                    System.out.printf("✂️  Barber is cutting hair for Customer %d%n", customerId);

                    // Simulate cutting hair (2-5 seconds)
                    Thread.sleep((random.nextInt(3) + 2) * 1000);

                    System.out.printf("✅ Barber finished cutting hair for Customer %d%n", customerId);
                } else {
                    System.out.println("🏠 Barber shop is closing (no customers for too long)");
                    isOpen = false;
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
        shutdownLatch.countDown();
    }

    public void addCustomer(int customerId) {
        executor.submit(() -> customerArrival(customerId));
    }

    private void customerArrival(int customerId) {
        System.out.printf("🚶 Customer %d is approaching the barber shop%n", customerId);

        try {
            // Try to enter the waiting room (non-blocking)
            boolean seated = waitingRoom.offer(customerId);

            if (seated) {
                System.out.printf("🪑 Customer %d is waiting (chairs occupied: %d)%n",
                    customerId, waitingRoom.size());
                System.out.printf("😊 Customer %d got haircut and left happy%n", customerId);
            } else {
                System.out.printf("😞 Customer %d left (waiting room is full)%n", customerId);
            }
        } catch (Exception e) {
            System.out.printf("❌ Customer %d encountered an error: %s%n", customerId, e.getMessage());
        }
    }

    public void shutdown() throws InterruptedException {
        isOpen = false;

        // Wait for barber to finish
        shutdownLatch.await(15, TimeUnit.SECONDS);

        executor.shutdown();
        if (!executor.awaitTermination(5, TimeUnit.SECONDS)) {
            executor.shutdownNow();
        }

        System.out.println("👋 Barber shop closed");
    }

    public static void main(String[] args) throws InterruptedException {
        int waitingChairs = 3;
        System.out.printf("🏪 Opening the Sleeping Barber Shop with %d waiting chairs%n", waitingChairs);

        SleepingBarber shop = new SleepingBarber(waitingChairs);
        Random random = new Random();

        // Start the barber
        shop.startBarber();

        // Generate customers at random intervals
        int customerCount = 0;
        long startTime = System.currentTimeMillis();
        long simulationDuration = 20000; // 20 seconds

        while (System.currentTimeMillis() - startTime < simulationDuration) {
            // Random chance of customer arriving (70% probability)
            if (random.nextFloat() < 0.7) {
                customerCount++;
                shop.addCustomer(customerCount);
            }

            Thread.sleep(800); // Wait 800ms between potential arrivals
        }

        System.out.println("\n🕐 Simulation time ended");
        shop.shutdown();
    }
}