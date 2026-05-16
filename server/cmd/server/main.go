package main

import (
	"log"
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}