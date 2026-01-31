package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"LoadBalancer/internal/pool"
)

func loadBalancerHandler(p *pool.ServerPool, w http.ResponseWriter, r *http.Request) {
	peer := p.GetNextPeer()
	log.Printf("Forwarding request to %s", peer.Host)

	proxy := httputil.NewSingleHostReverseProxy(peer)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[%s] connection failed: %v", peer.Host, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service unavailable"))
	}

	proxy.ServeHTTP(w, r)
}

func main() {
	servers := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	p := pool.NewServerPool(servers)

	server := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loadBalancerHandler(p, w, r)
		}),
	}

	log.Printf("Starting load balancer on port %s", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}
