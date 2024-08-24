package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr        string
	Weight      int
	Connections int64
}

type ServerPool struct {
	servers sync.Map
}

func (sp *ServerPool) AddServer(addr string, weight int) {
	sp.servers.Store(addr, &Server{Addr: addr, Weight: weight})
}

func (sp *ServerPool) RemoveServer(addr string) {
	sp.servers.Delete(addr)
}

func (sp *ServerPool) GetServer() []*Server {
	servers := make([]*Server, 0)
	sp.servers.Range(func(_, value any) bool {
		servers = append(servers, value.(*Server))
		return true
	})
	return servers
}

func LeastConnections(servers []*Server) *Server {
	var bestServer *Server
	leastConns := int64(1<<63 - 1)
	for _, server := range servers {
		if server.Connections < leastConns {
			bestServer = server
			leastConns = server.Connections
		}
	}
	return bestServer
}

func healthCheck(server *Server) bool {
	url, err := url.Parse("http://" + server.Addr)
	if err != nil {
		return false
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func runHealthCheck(sp *ServerPool, interval time.Duration) {
	for {
		time.Sleep(interval)
		sp.servers.Range(func(addr, server any) bool {
			healthy := healthCheck(server.(*Server))
			if !healthy {
				sp.RemoveServer(addr.(string))
			}
			return true
		})
	}
}

type LoadBalancer struct {
	serverPool *ServerPool
	algorithm  func(server []*Server) *Server
	interval   time.Duration
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servers := lb.serverPool.GetServer()
	if len(servers) == 0 {
		http.Error(w, "there are no servers available", http.StatusServiceUnavailable)
		return
	}

	server := lb.algorithm(servers)
	if server == nil {
		http.Error(w, "there are no servers available", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&server.Connections, 1)
	defer atomic.AddInt64(&server.Connections, -1)

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   server.Addr,
	})
	proxy.ServeHTTP(w, r)
}

func main() {
	serverPool := &ServerPool{}
	algolithm := LeastConnections
	interval := 5 * time.Second

	serverPool.AddServer("127.0.0.1:8081", 1)
	serverPool.AddServer("127.0.0.1:8082", 2)
	serverPool.AddServer("127.0.0.1:8083", 3)

	lb := &LoadBalancer{
		serverPool: serverPool,
		algorithm:  algolithm,
		interval:   interval,
	}

	go runHealthCheck(serverPool, interval)

	http.Handle("/", lb)
	fmt.Println("starting load balancer server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
