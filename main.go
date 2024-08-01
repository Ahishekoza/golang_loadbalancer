package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type SimpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

// Method with returns SimpleServer is used create SimpleServers
func NewSimpleServer(addr string) *SimpleServer {
	fmt.Println("Address ", addr)
	serverURL, err := url.Parse(addr)

	fmt.Println("Server URL ", serverURL)

	if err != nil {
		fmt.Println("error", err.Error())
	}

	return &SimpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}
}

// Serve Method will pass the client request to the server
func (s *SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

type LoadBalancer struct {
	servers         []*SimpleServer
	roundRobinIndex uint32
}

// Method NewLoadBalancer which returns a load balancer is use to create new load balancers
func NewLoadBalancer(servers []*SimpleServer) *LoadBalancer {
	return &LoadBalancer{
		servers: servers,
	}
}

// --It will  give me the next available server
func (lb *LoadBalancer) getNextServer() *SimpleServer {
	index := atomic.AddUint32(&lb.roundRobinIndex, 1) - 1
	return lb.servers[index%uint32(len(lb.servers))]
}

// Client will send a request to the server proxy and serveproxy will first check which load balancer is available
// then the request is forwarded to the available server
func (lb *LoadBalancer) ServerProxy(rw http.ResponseWriter, req *http.Request) {
	server := lb.getNextServer()
	fmt.Println("Avaliable Server", server)
	server.Serve(rw, req)
}

func main() {
	servers := []*SimpleServer{
		NewSimpleServer("https://www.google.com"),
		NewSimpleServer("http://www.duckduckgo.com"),
		NewSimpleServer("https://www.digitalocean.com"),
	}

	fmt.Println(servers)

	lb := NewLoadBalancer(servers)

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		lb.ServerProxy(rw, req)
	})

	http.ListenAndServe(":8002", nil)
}
