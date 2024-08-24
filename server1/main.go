package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	homeHandler := func(w http.ResponseWriter, r *http.Request) {
		l := log.New(os.Stdout, "[Server-1]", log.Ldate|log.Ltime)
		l.Printf("running...")
		io.WriteString(w, "hello world from server 1\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("starting server-1 at port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
