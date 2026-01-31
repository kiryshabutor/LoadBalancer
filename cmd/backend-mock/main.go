package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func launchMiniServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		log.Printf("Service on port %s received request: %s", port, string(body))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Request from service on port %s\n", port)
	})

	log.Printf("Starting mini server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	launchMiniServer(port)
}
