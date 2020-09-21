package main

import (
	"io"
	"net"
	"net/http"
	"time"
)

type Proxy struct {
	http.Handler
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleHTTPS(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	_, _ = io.Copy(destination, source)
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	destinationConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking error", http.StatusInternalServerError)
		return
	}
	sourceConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(destinationConn, sourceConn)
	go transfer(sourceConn, destinationConn)
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer res.Body.Close()

	for header, value := range res.Header {
		w.Header()[header] = value
	}
	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}
