import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.locks.ReentrantLock;
import java.util.concurrent.locks.Condition;
import java.util.HashSet;
import java.util.Set;

public class CustomThreadPoolExecutor implements Executor {
    private final int corePoolSize;
    private final int maximumPoolSize;
    private final long keepAliveTime;
    private final TimeUnit unit;
    private final BlockingQueue<Runnable> workQueue;
    private final ThreadFactory threadFactory;
    private final RejectedExecutionHandler handler;

    private final AtomicInteger poolSize = new AtomicInteger(0);
    private final AtomicInteger activeCount = new AtomicInteger(0);
    private final Set<Worker> workers = new HashSet<>();
    private final ReentrantLock mainLock = new ReentrantLock();

    private volatile boolean shutdown = false;
    private final Condition termination = mainLock.newCondition();

    public CustomThreadPoolExecutor(int corePoolSize, int maximumPoolSize, long keepAliveTime,
                                   TimeUnit unit, BlockingQueue<Runnable> workQueue) {
        this(corePoolSize, maximumPoolSize, keepAliveTime, unit, workQueue,
             Executors.defaultThreadFactory(), new AbortPolicy());
    }

    public CustomThreadPoolExecutor(int corePoolSize, int maximumPoolSize, long keepAliveTime,
                                   TimeUnit unit, BlockingQueue<Runnable> workQueue,
                                   ThreadFactory threadFactory, RejectedExecutionHandler handler) {
        if (corePoolSize < 0 || maximumPoolSize <= 0 || maximumPoolSize < corePoolSize || keepAliveTime < 0)
            throw new IllegalArgumentException();
        if (workQueue == null || threadFactory == null || handler == null)
            throw new NullPointerException();

        this.corePoolSize = corePoolSize;
        this.maximumPoolSize = maximumPoolSize;
        this.keepAliveTime = keepAliveTime;
        this.unit = unit;
        this.workQueue = workQueue;
        this.threadFactory = threadFactory;
        this.handler = handler;
    }

    @Override
    public void execute(Runnable command) {
        if (command == null)
            throw new NullPointerException();

        int c = poolSize.get();

        if (c < corePoolSize) {
            if (addWorker(command, true))
                return;
            c = poolSize.get();
        }

        if (!shutdown && workQueue.offer(command)) {
            if (shutdown && !workQueue.remove(command)) {
                return;
            }
            if (poolSize.get() == 0)
                addWorker(null, false);
        } else if (!addWorker(command, false)) {
            handler.rejectedExecution(command, null);
        }
    }

    private boolean addWorker(Runnable firstTask, boolean core) {
        retry:
        for (;;) {
            int c = poolSize.get();

            if (shutdown)
                return false;

            int wc = c;
            if (wc >= ((core ? corePoolSize : maximumPoolSize) & 0x1fffffff))
                return false;

            if (poolSize.compareAndSet(c, c + 1))
                break retry;
        }

        boolean workerStarted = false;
        boolean workerAdded = false;
        Worker w = null;
        try {
            w = new Worker(firstTask);
            final Thread t = w.thread;
            if (t != null) {
                final ReentrantLock mainLock = this.mainLock;
                mainLock.lock();
                try {
                    if (!shutdown) {
                        workers.add(w);
                        workerAdded = true;
                        int s = workers.size();
                    }
                } finally {
                    mainLock.unlock();
                }
                if (workerAdded) {
                    t.start();
                    workerStarted = true;
                }
            }
        } finally {
            if (!workerStarted)
                addWorkerFailed(w);
        }
        return workerStarted;
    }

    private void addWorkerFailed(Worker w) {
        final ReentrantLock mainLock = this.mainLock;
        mainLock.lock();
        try {
            if (w != null)
                workers.remove(w);
            poolSize.decrementAndGet();
        } finally {
            mainLock.unlock();
        }
    }

    private void processWorkerExit(Worker w, boolean completedAbruptly) {
        if (completedAbruptly)
            poolSize.decrementAndGet();

        final ReentrantLock mainLock = this.mainLock;
        mainLock.lock();
        try {
            workers.remove(w);
            poolSize.decrementAndGet();
            termination.signalAll();
        } finally {
            mainLock.unlock();
        }

        if (!completedAbruptly) {
            int min = 0;
            if (!shutdown)
                min = corePoolSize;
            if (min == 0 && !workQueue.isEmpty())
                min = 1;
            if (poolSize.get() >= min)
                return;
        }
        addWorker(null, false);
    }

    private Runnable getTask() {
        boolean timedOut = false;

        for (;;) {
            if (shutdown && workQueue.isEmpty())
                return null;

            int wc = poolSize.get();
            boolean timed = wc > corePoolSize;

            if ((wc > maximumPoolSize || (timed && timedOut)) && (wc > 1 || workQueue.isEmpty())) {
                if (poolSize.compareAndSet(wc, wc - 1))
                    return null;
                continue;
            }

            try {
                Runnable r = timed ? workQueue.poll(keepAliveTime, unit) : workQueue.take();
                if (r != null)
                    return r;
                timedOut = true;
            } catch (InterruptedException retry) {
                timedOut = false;
            }
        }
    }

    final void runWorker(Worker w) {
        Thread wt = Thread.currentThread();
        Runnable task = w.firstTask;
        w.firstTask = null;
        boolean completedAbruptly = true;
        try {
            while (task != null || (task = getTask()) != null) {
                activeCount.incrementAndGet();
                try {
                    beforeExecute(wt, task);
                    Throwable thrown = null;
                    try {
                        task.run();
                    } catch (RuntimeException x) {
                        thrown = x; throw x;
                    } catch (Error x) {
                        thrown = x; throw x;
                    } catch (Throwable x) {
                        thrown = x; throw new Error(x);
                    } finally {
                        afterExecute(task, thrown);
                    }
                } finally {
                    task = null;
                    activeCount.decrementAndGet();
                }
            }
            completedAbruptly = false;
        } finally {
            processWorkerExit(w, completedAbruptly);
        }
    }

    protected void beforeExecute(Thread t, Runnable r) { }
    protected void afterExecute(Runnable r, Throwable t) { }

    public void shutdown() {
        final ReentrantLock mainLock = this.mainLock;
        mainLock.lock();
        try {
            shutdown = true;
            interruptIdleWorkers();
        } finally {
            mainLock.unlock();
        }
    }

    private void interruptIdleWorkers() {
        for (Worker w : workers) {
            w.interruptIfStarted();
        }
    }

    public boolean isShutdown() {
        return shutdown;
    }

    public boolean awaitTermination(long timeout, TimeUnit unit) throws InterruptedException {
        long nanos = unit.toNanos(timeout);
        final ReentrantLock mainLock = this.mainLock;
        mainLock.lock();
        try {
            for (;;) {
                if (workers.isEmpty())
                    return true;
                if (nanos <= 0)
                    return false;
                nanos = termination.awaitNanos(nanos);
            }
        } finally {
            mainLock.unlock();
        }
    }

    public int getPoolSize() {
        return poolSize.get();
    }

    public int getActiveCount() {
        return activeCount.get();
    }

    public int getQueueSize() {
        return workQueue.size();
    }

    private final class Worker implements Runnable {
        final Thread thread;
        Runnable firstTask;

        Worker(Runnable firstTask) {
            this.firstTask = firstTask;
            this.thread = threadFactory.newThread(this);
        }

        public void run() {
            runWorker(this);
        }

        void interruptIfStarted() {
            Thread t;
            if ((t = thread) != null && !t.isInterrupted()) {
                try {
                    t.interrupt();
                } catch (SecurityException ignore) {
                }
            }
        }
    }

    public static class AbortPolicy implements RejectedExecutionHandler {
        public void rejectedExecution(Runnable r, ThreadPoolExecutor e) {
            throw new RejectedExecutionException("Task " + r.toString() + " rejected");
        }
    }
}