import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class ScheduledExecutorDemo {
    public static void main(String[] args) throws InterruptedException {
        CustomScheduledExecutor executor = new CustomScheduledExecutor(2);

        AtomicInteger counter = new AtomicInteger(0);

        System.out.println("Starting scheduled task demo...");

        ScheduledFuture<?> future = executor.scheduleAtFixedRate(() -> {
            int count = counter.incrementAndGet();
            System.out.println("Task execution #" + count + " at " + System.currentTimeMillis());

            if (count >= 5) {
                System.out.println("Task completed 5 executions");
            }
        }, 1, 2, TimeUnit.SECONDS);

        Thread.sleep(12000);

        System.out.println("Cancelling task...");
        future.cancel(true);

        System.out.println("Shutting down executor...");
        executor.shutdown();

        System.out.println("Demo completed");
    }
}