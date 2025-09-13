import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class ThreadPoolExecutorDemo {
    public static void main(String[] args) throws InterruptedException {

        CustomThreadPoolExecutor executor = new CustomThreadPoolExecutor(
            2,                              // core pool size
            4,                              // maximum pool size
            60L,                            // keep alive time
            TimeUnit.SECONDS,               // time unit
            new LinkedBlockingQueue<>(10)   // work queue with capacity 10
        );

        AtomicInteger taskCounter = new AtomicInteger(0);

        System.out.println("Starting ThreadPool demo...");
        System.out.println("Core pool size: 2, Max pool size: 4, Queue capacity: 10");

        // Submit 15 tasks to demonstrate pool behavior
        for (int i = 1; i <= 15; i++) {
            final int taskId = i;
            try {
                executor.execute(() -> {
                    int taskNum = taskCounter.incrementAndGet();
                    String threadName = Thread.currentThread().getName();
                    System.out.printf("Task %d (ID: %d) started on thread: %s%n",
                                    taskNum, taskId, threadName);

                    try {
                        Thread.sleep(2000); // Simulate work
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        return;
                    }

                    System.out.printf("Task %d (ID: %d) completed on thread: %s%n",
                                    taskNum, taskId, threadName);
                });

                System.out.printf("Submitted task %d - Pool size: %d, Active: %d, Queue size: %d%n",
                                i, executor.getPoolSize(), executor.getActiveCount(), executor.getQueueSize());

            } catch (RejectedExecutionException e) {
                System.out.printf("Task %d REJECTED: %s%n", i, e.getMessage());
            }

            Thread.sleep(200); // Small delay between submissions
        }

        System.out.println("\nWaiting for tasks to complete...");
        Thread.sleep(10000);

        System.out.println("\nShutting down executor...");
        executor.shutdown();

        if (executor.awaitTermination(10, TimeUnit.SECONDS)) {
            System.out.println("All tasks completed successfully");
        } else {
            System.out.println("Timeout waiting for tasks to complete");
        }

        System.out.println("Demo completed");
    }
}