import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class RWLock1 {

    private boolean isWriterActive = false;
    private int readerCount = 0;
    private final Lock lock = new ReentrantLock();
    private final Condition readCond = lock.newCondition();
    private final Condition writeCond = lock.newCondition();

    public RWLock1() {}

    public void rlock() throws InterruptedException {
        lock.lock();
        try {
            while (isWriterActive) {
                readCond.await();
            }
            readerCount++;
        } finally {
            lock.unlock();
        }
    }

    public void runlock() {
        lock.lock();
        try {
            readerCount--;
            if (readerCount == 0) {
                writeCond.signal();
            }
        } finally {
            lock.unlock();
        }
    }

    public void wlock() throws InterruptedException {
        lock.lock();
        try {
            while (readerCount > 0 || isWriterActive) {
                writeCond.await();
            }
            isWriterActive = true;
        } finally {
            lock.unlock();
        }
    }

    public void wunlock() {
        lock.lock();
        try {
            isWriterActive = false;
            readCond.signalAll();
            writeCond.signal();
        } finally {
            lock.unlock();
        }
    }
}