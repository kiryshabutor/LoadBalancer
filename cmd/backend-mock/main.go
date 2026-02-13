package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func launchMiniServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		defer r.Body.Close()

		// Simulate 10-30ms latency (typical DB call)
		time.Sleep(20 * time.Millisecond)

		// Simulate some CPU work
		h := sha256.New()
		for i := 0; i < 1000; i++ {
			h.Write([]byte(fmt.Sprintf("%d", i)))
		}

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
