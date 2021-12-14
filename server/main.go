package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Example http server
type serve struct{}

var start time.Time

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	elapsed := time.Since(start)
	fmt.Fprintf(w, "Serve %s\n", elapsed)
	fmt.Printf("Serve %s\n", elapsed)
}

func init() {
	fmt.Println("I am init")
}

func main() {
	http.HandleFunc("/", ServeHTTP)
	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
