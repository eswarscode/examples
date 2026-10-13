package main
import "fmt"
import "net/http"


func main() {
	fmt.Println("client server")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /get", GetHandler)
	svr := http.Server {Addr : ":8085", Handler : mux,}
	svr.ListenAndServe()

}

func GetHandler(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("hi get"))
}