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
		l := log.New(os.Stdout, "[Server-2]", log.Ldate|log.Ltime)
		l.Printf("running...")
		io.WriteString(w, "hello world from server 2\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("starting server-2 at port 8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
