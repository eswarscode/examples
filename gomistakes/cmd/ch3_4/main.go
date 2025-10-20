package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"time"
)

func main() {
	//testSlicesAppend()
	//testMemAlloc()
	testDeferInLoop()
}

func testSlicesAppend() {
	var mySlice []int = make([]int, 5, 10)
	var initialSlice []int = []int{0, 1, 2, 3, 4}

	copy(mySlice, initialSlice)

	newSlice := mySlice[3:4]       // default capacity is as myslice
	capSlice := mySlice[3:4:7]     // set capacity up to index 7
	newSlice = append(newSlice, 5) // this will modify mySlice as well
	fmt.Println(newSlice, "cap is", cap(newSlice), "len is", len(newSlice))
	fmt.Println(mySlice, "cap is", cap(mySlice), "len is", len(mySlice))
	fmt.Println(capSlice, "cap is", cap(capSlice), "len is", len(capSlice))

}

func testMemAlloc() {
	bytes := getMemSlices()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	fmt.Println("After GC:", len(bytes))
	printAlloc() // till you have reference from original slice, memory won't be freed
}
func testDeferInLoop() {
	var filesch <-chan string
	var done chan struct{} = make(chan struct{})
	go func() {

		ch := make(chan string)
		filesch = ch
		done <- struct{}{} // Signal the main goroutine

		defer close(ch)

		ch <- "conf/file1.txt"
		ch <- "conf/file2.txt"
	}()

	// This will open all files and defer their closing till end of function
	// for path := range filesch {
	//         file := os.Open(path)
	// 		if file != nil {
	// 			printFile(file)
	// 			defer file.Close()  // this will delay closing of all files till end of function
	// 		}
	// }
	<-done
	for file := range filesch {
		printFile(file) // this will close file before next iteration
	}

}

func printAlloc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d KB\n", m.Alloc/1024)
}

func getMemSlices() [][]byte {
	results := make([][]byte, 100, 100)
	for i := 0; i < 100; i++ {
		results[i] = make([]byte, 1024*1024) // 1 MB
	}
	fmt.Println("After allocating 100 MB:")
	printAlloc()
	return results[0:1]
}

func printFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
