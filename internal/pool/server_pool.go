package pool

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	servers []*Backend
	current uint64
	mu      sync.RWMutex
}

type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
	Proxy *httputil.ReverseProxy
}

func NewServerPool(servers []string) *ServerPool {
	pool := &ServerPool{
		servers: make([]*Backend, len(servers)),
		current: 0,
	}
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 250,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}

	for i, s := range servers {
		u, err := url.Parse(s)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.Transport = transport

		pool.servers[i] = &Backend{
			URL:   u,
			Proxy: proxy,
		}
		pool.servers[i].Alive.Store(true)

		backend := pool.servers[i]
		pool.servers[i].Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[%s] connection failed: %v", backend.URL.Host, err)
			pool.MarkServerDown(backend)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service unavailable"))
		}
	}
	return pool
}

func (s *ServerPool) AddPeer(server string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	s.servers = append(s.servers, &Backend{
		URL:   u,
		Proxy: httputil.NewSingleHostReverseProxy(u),
	})
	s.servers[len(s.servers)-1].Alive.Store(true)
}

func (s *ServerPool) GetNextBackend() *Backend {
	length := uint64(len(s.servers))
	for i := uint64(0); i < length; i++ {
		next := atomic.AddUint64(&s.current, 1) % length
		if s.servers[next].Alive.Load() {
			return s.servers[next]
		}
	}
	return nil
}

func (s *ServerPool) MarkServerDown(backend *Backend) {
	if backend.Alive.Swap(false) {
		log.Printf("Server %s is down", backend.URL.Host)
		go s.StartHealthCheck(backend)
	}
}

func (s *ServerPool) StartHealthCheck(backend *Backend) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		conn, err := net.DialTimeout("tcp", backend.URL.Host, 1*time.Second)
		if err == nil {
			conn.Close()
			s.MarkServerUp(backend)
			return
		}
	}
}

func (s *ServerPool) MarkServerUp(backend *Backend) {
	if !backend.Alive.Swap(true) {
		log.Printf("Server %s is up", backend.URL.Host)
	}
}
