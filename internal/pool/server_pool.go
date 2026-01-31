package pool

import (
	"log"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	servers []*url.URL
	current uint64
}

func NewServerPool(servers []string) *ServerPool {
	pool := &ServerPool{
		servers: make([]*url.URL, len(servers)),
		current: 0,
	}
	for i, s := range servers {
		u, err := url.Parse(s)
		if err != nil {
			log.Fatal(err)
		}
		pool.servers[i] = u
	}
	return pool
}

func (s *ServerPool) AddPeer(server string) {
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	s.servers = append(s.servers, u)
}

func (s *ServerPool) GetNextPeer() *url.URL {
	next := atomic.AddUint64(&s.current, 1) % uint64(len(s.servers))
	return s.servers[next]
}
