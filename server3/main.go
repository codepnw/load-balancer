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
		l := log.New(os.Stdout, "[Server-3]", log.Ldate|log.Ltime)
		l.Printf("running...")
		io.WriteString(w, "hello world from server 3\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("starting server-3 at port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
