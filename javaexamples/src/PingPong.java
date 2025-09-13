import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;

public class PingPong {
    private final BlockingQueue<String> pingChannel;
    private final BlockingQueue<String> pongChannel;

    public PingPong(int capacity) {
        this.pingChannel = new LinkedBlockingQueue<>(capacity);
        this.pongChannel = new LinkedBlockingQueue<>(capacity);
    }

    // Pinger sends "ping" and waits for "pong"
    public void pinger() {
        while (!Thread.currentThread().isInterrupted()) {
            try {
                // Send "ping"
                pingChannel.put("ping");
                System.out.println("Ping!");

                // Wait for "pong"
                pongChannel.take();

            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
    }

    // Ponger waits for "ping" and sends "pong"
    public void ponger() {
        while (!Thread.currentThread().isInterrupted()) {
            try {
                // Wait for "ping"
                pingChannel.take();

                // Send "pong"
                pongChannel.put("pong");
                System.out.println("Pong!");

                // Sleep for 180 seconds (equivalent to Go version)
                Thread.sleep(TimeUnit.SECONDS.toMillis(180));

            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
    }

    public static void main(String[] args) {
        PingPong game = new PingPong(5);

        // Start pinger and ponger as separate threads
        Thread pingerThread = new Thread(game::pinger, "Pinger");
        Thread pongerThread = new Thread(game::ponger, "Ponger");

        pingerThread.start();
        pongerThread.start();

        // Keep main thread alive (equivalent to select{} in Go)
        try {
            pingerThread.join();
            pongerThread.join();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}