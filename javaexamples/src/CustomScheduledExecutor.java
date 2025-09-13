import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicLong;

public class CustomScheduledExecutor {
    private final ThreadPoolExecutor executor;
    private final AtomicBoolean shutdown = new AtomicBoolean(false);

    public CustomScheduledExecutor(int corePoolSize) {
        this.executor = new ThreadPoolExecutor(
            corePoolSize,
            corePoolSize,
            0L,
            TimeUnit.MILLISECONDS,
            new LinkedBlockingQueue<>(),
            new ThreadFactory() {
                private final AtomicLong threadNumber = new AtomicLong(1);

                @Override
                public Thread newThread(Runnable r) {
                    Thread t = new Thread(r, "CustomScheduledExecutor-" + threadNumber.getAndIncrement());
                    t.setDaemon(false);
                    return t;
                }
            }
        );
    }

    public ScheduledFuture<?> scheduleAtFixedRate(Runnable command, long initialDelay, long period, TimeUnit unit) {
        if (command == null || unit == null) {
            throw new NullPointerException();
        }
        if (period <= 0) {
            throw new IllegalArgumentException("Period must be positive");
        }

        return new FixedRateScheduledFuture(command, initialDelay, period, unit);
    }

    public void shutdown() {
        shutdown.set(true);
        executor.shutdown();
    }

    public boolean isShutdown() {
        return shutdown.get();
    }

    private class FixedRateScheduledFuture implements ScheduledFuture<Object> {
        private final Runnable command;
        private final long periodNanos;
        private final AtomicBoolean cancelled = new AtomicBoolean(false);
        private volatile boolean done = false;
        private volatile long nextExecutionTime;

        public FixedRateScheduledFuture(Runnable command, long initialDelay, long period, TimeUnit unit) {
            this.command = command;
            this.periodNanos = unit.toNanos(period);
            this.nextExecutionTime = System.nanoTime() + unit.toNanos(initialDelay);

            scheduleNext();
        }

        private void scheduleNext() {
            if (cancelled.get() || shutdown.get()) {
                done = true;
                return;
            }

            long delay = nextExecutionTime - System.nanoTime();
            if (delay <= 0) {
                delay = 0;
            }

            executor.schedule(() -> {
                if (cancelled.get() || shutdown.get()) {
                    done = true;
                    return;
                }

                long startTime = System.nanoTime();
                try {
                    command.run();
                } catch (Exception e) {
                    System.err.println("Exception in scheduled task: " + e.getMessage());
                    e.printStackTrace();
                }

                if (!cancelled.get() && !shutdown.get()) {
                    nextExecutionTime = startTime + periodNanos;
                    scheduleNext();
                }
            }, delay, TimeUnit.NANOSECONDS);
        }

        @Override
        public long getDelay(TimeUnit unit) {
            return unit.convert(nextExecutionTime - System.nanoTime(), TimeUnit.NANOSECONDS);
        }

        @Override
        public int compareTo(Delayed other) {
            if (other == this) return 0;
            long diff = getDelay(TimeUnit.NANOSECONDS) - other.getDelay(TimeUnit.NANOSECONDS);
            return (diff < 0) ? -1 : (diff > 0) ? 1 : 0;
        }

        @Override
        public boolean cancel(boolean mayInterruptIfRunning) {
            return cancelled.compareAndSet(false, true);
        }

        @Override
        public boolean isCancelled() {
            return cancelled.get();
        }

        @Override
        public boolean isDone() {
            return done || cancelled.get();
        }

        @Override
        public Object get() throws InterruptedException, ExecutionException {
            throw new UnsupportedOperationException("get() not supported for repeating tasks");
        }

        @Override
        public Object get(long timeout, TimeUnit unit) throws InterruptedException, ExecutionException, TimeoutException {
            throw new UnsupportedOperationException("get() not supported for repeating tasks");
        }
    }
}