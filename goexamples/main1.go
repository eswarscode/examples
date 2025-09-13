package main

import "fmt"
import "net/http"

func helloHandler1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	fmt.Println("hi")
	http.HandleFunc("/hi", helloHandler1)
	http.ListenAndServe(":8080", nil)

}
