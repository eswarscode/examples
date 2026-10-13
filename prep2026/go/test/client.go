package main
import "fmt"
import "net/http"
import "io"

func main() {
	ClientFunc()
	GetFunc()
}
func ClientFunc() {
	client := http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8085/get", nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	str, _ := io.ReadAll(resp.Body)
	fmt.Println(string(str))
	resp.Body.Close()
}

func GetFunc() {
	
	resp, err := http.Get("http://localhost:8085/get")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	str, _ := io.ReadAll(resp.Body)
	fmt.Println(string(str))
	resp.Body.Close()
}