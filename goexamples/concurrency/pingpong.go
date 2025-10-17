package main

import (
	"fmt"
	"time"
)

// pinger sends "ping" and waits for "pong"
func pinger(pingChan chan<- string, pongChan <-chan string) {
	for {
		// Send "ping"
		pingChan <- "ping"
		fmt.Println("Ping!")

		// Wait for "pong"
		<-pongChan

		//time.Sleep(time.Duration(rand.Intn(5)+1) * time.Millisecond)

	}
}

// ponger waits for "ping" and sends "pong"
func ponger(pingChan <-chan string, pongChan chan<- string) {
	for {
		// Wait for "ping"
		<-pingChan
		//fmt.Println("Received Ping!")

		// Send "pong"
		pongChan <- "pong"
		fmt.Println("Pong!")

		//time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		time.Sleep(180 * time.Second)
	}
}

func main() {
	// Create channels for communication
	pingChannel := make(chan string, 5)
	pongChannel := make(chan string, 5)

	// Start the pinger and ponger as goroutines
	go pinger(pingChannel, pongChannel)
	go ponger(pingChannel, pongChannel)

	// Keep the main goroutine alive to allow the other goroutines to run
	select {} // This will block indefinitely, allowing pinger and ponger to run
}
