package main

import (
	"fmt"
	"os"
	"bufio"
)



func main() {
	fmt.Println("main \n ")
	FileHandler2()
}


func FileHandler1() {
	data, err := os.ReadFile("1.text")
	if err != nil {
		fmt.Println(err)
		return
	}

	str := string(data)
	fmt.Println(str)

}

func FileHandler2() {
	file, err := os.Open("1.text")
	
	if err != nil {
		return
	}

	sc  := bufio.NewScanner(file)
	
	for  sc.Scan()  {
            fmt.Println("hi --- " + sc.Text())
	}

	
	fmt.Println(sc.Err())
	defer file.Close()

}